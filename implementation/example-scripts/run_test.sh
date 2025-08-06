#!/bin/bash

# --- Configuration ---
APP_SERVER_DIR="app-server"
MONITOR_DIR="resource-monitor"
LOAD_GEN_DIR="load-generator"
LOG_FILE="resource_usage.log"
K6_SUMMARY_FILE="k6_summary.txt"
BIN_DIR="bin"

# --- Functions ---

# Function to print a formatted header
print_header() {
    echo "======================================================================"
    echo "  $1"
    echo "======================================================================"
}

# Function to clean up background processes
cleanup() {
    print_header "CLEANUP"
    echo "Stopping all background processes..."
    # Kill the process group to ensure all children are terminated
    if [ -n "$MONITOR_PID" ]; then
        kill -9 "$MONITOR_PID" 2>/dev/null
        echo "Stopped Resource Monitor (PID: $MONITOR_PID)"
    fi
    if [ -n "$APP_SERVER_PID" ]; then
        kill -9 "$APP_SERVER_PID" 2>/dev/null
        echo "Stopped App Server (PID: $APP_SERVER_PID)"
    fi
    # Don't remove the log file here, it's handled at the end
    rm -f "$K6_SUMMARY_FILE" "app-server.log" "run_test.log"
    echo "Cleanup complete."
}

# Trap EXIT signal to ensure cleanup runs
trap cleanup EXIT

# --- Main Script ---

if [ -f "$LOG_FILE" ]; then
    print_header "REUSING EXISTING LOG FILE"
    echo "Log file '$LOG_FILE' found. Skipping test run and analyzing existing data."
else
    # 0. Setup Local Tools
    print_header "SETTING UP LOCAL TOOLS"
    ./setup_tools.sh || { echo "Failed to setup tools"; exit 1; }

    # 1. Build Applications
    print_header "BUILDING APPLICATIONS"
    echo "Building app-server..."
    (cd "$APP_SERVER_DIR" && go build -o app-server .) || { echo "Failed to build app-server"; exit 1; }
    echo "Building resource-monitor..."
    (cd "$MONITOR_DIR" && go build -o resource-monitor .) || { echo "Failed to build resource-monitor"; exit 1; }
    echo "Build complete."

    # 2. Start App Server
    print_header "STARTING APP SERVER"
    # First, ensure port 8080 is free
    echo "Checking for existing process on port 8080..."
    # Use netstat to find the listening process and awk/cut to extract the PID
    EXISTING_PID=$(netstat -tulpn | grep ':8080' | awk '{print $7}' | cut -d'/' -f1)
    if [ -n "$EXISTING_PID" ]; then
        echo "Found existing process $EXISTING_PID on port 8080. Terminating it."
        kill -9 "$EXISTING_PID"
        sleep 1
    fi
    # Start the server in the background with high load
    echo "Starting app-server with LOAD_CPU_ITERATIONS=0 and LOAD_MEM_MB=1"
    LOAD_CPU_ITERATIONS=0 LOAD_MEM_MB=1 ./"$APP_SERVER_DIR"/app-server > app-server.log 2>&1 &
    APP_SERVER_PID=$(pgrep -f app-server)
    # Wait a moment for the server to initialize
    sleep 2
    # Check if the server is running
    if ! ps -p "$APP_SERVER_PID" > /dev/null; then
        echo "Failed to start app-server."
        exit 1
    fi
    echo "App Server started with PID: $APP_SERVER_PID"

    # 3. Start Resource Monitor
    print_header "STARTING RESOURCE MONITOR"
    ./"$MONITOR_DIR"/resource-monitor "$APP_SERVER_PID" > "$LOG_FILE" &
    MONITOR_PID=$!
    sleep 1
    echo "Resource Monitor started with PID: $MONITOR_PID, logging to $LOG_FILE"

    # 4. Run Load Test
    print_header "RUNNING k6 LOAD TEST (1000 RPS for 10s)"
    (cd "$LOAD_GEN_DIR" && ../bin/k6 run k6-script.js --summary-export="../$K6_SUMMARY_FILE") > run_test.log 2>&1
    K6_EXIT_CODE=$?
    if [ $K6_EXIT_CODE -ne 0 ]; then
        echo "k6 load test failed with exit code $K6_EXIT_CODE."
        # Don't exit immediately, allow results to be processed
    fi
    cat run_test.log | ./bin/filter_k6_output.awk

    # Give the resource monitor a moment to log the final stats before killing it
    sleep 2

    # Stop the app server and monitor now that the test is done
    kill -9 "$MONITOR_PID" 2>/dev/null
    kill -9 "$APP_SERVER_PID" 2>/dev/null
fi

# 5. Process and Display Results
print_header "TEST RESULTS"

# Process Resource Usage
echo "--- Resource Usage Summary ---"
if [ ! -f "$LOG_FILE" ]; then
    echo "Log file '$LOG_FILE' not found. Cannot process resource usage."
else
    # Skip the header line of the CSV
    DATA_ROWS=$(tail -n +2 "$LOG_FILE")
    # Calculate average and peak values using awk
    AVG_CPU=$(echo "$DATA_ROWS" | awk -F, '{ total += $2; count++ } END { if (count > 0) printf "%.4f", total/count; else print "0"; }')
    PEAK_CPU=$(echo "$DATA_ROWS" | awk -F, 'BEGIN { max=0 } { if ($2>max) max=$2 } END { printf "%.4f", max }')
    AVG_MEM=$(echo "$DATA_ROWS" | awk -F, '{ total += $3; count++ } END { if (count > 0) printf "%.2f", total/count; else print "0"; }')
    PEAK_MEM=$(echo "$DATA_ROWS" | awk -F, 'BEGIN { max=0 } { if ($3>max) max=$3 } END { printf "%.2f", max }')

    echo "Average CPU Cores: $AVG_CPU"
    echo "Peak CPU Cores:    $PEAK_CPU"
    echo "Average Memory (MB): $AVG_MEM"
    echo "Peak Memory (MB):    $PEAK_MEM"
    echo ""
fi


# Process k6 Summary
echo "--- k6 Performance Summary ---"
if [ -f "$K6_SUMMARY_FILE" ]; then
    # Extract and format key metrics from the k6 JSON summary
    ./"$BIN_DIR"/jq -r '
        .metrics |
        {
            "HTTP Requests": .http_reqs.count,
            "RPS (actual)": .http_reqs.rate,
            "Failed Requests": .http_req_failed.value,
            "Request Duration (p95)": .http_req_duration."p(95)",
            "Request Duration (avg)": .http_req_duration.avg,
            "Browse Duration (p95)": .http_req_duration_browse."p(95)",
            "Browse Duration (avg)": .http_req_duration_browse.avg,
            "Submit Duration (p95)": .http_req_duration_submit."p(95)",
            "Submit Duration (avg)": .http_req_duration_submit.avg
        } |
        to_entries | .[] |
        if (.key | contains("Duration")) then
            "\(.key) (ms): \(.value)"
        else
            "\(.key): \(.value)"
        end
    ' "$K6_SUMMARY_FILE"
else
    echo "k6 summary file not found."
fi

# Erase log file on successful completion
if [ -f "$LOG_FILE" ]; then
    print_header "SUCCESS"
    echo "Analysis complete. Deleting log file '$LOG_FILE'."
    rm "$LOG_FILE"
fi


echo "======================================================================"
echo "                      TEST COMPLETED"
echo "======================================================================"