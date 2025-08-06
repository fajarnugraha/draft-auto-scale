#!/bin/bash

# --- Configuration ---
LOG_FILE="resource_usage.log"
K6_SUMMARY_FILE="k6_summary.txt"
LOAD_GEN_DIR="load-generator"
BIN_DIR="bin"

# --- Functions ---

print_header() {
    echo "======================================================================"
    echo "  $1"
    echo "======================================================================"
}

cleanup() {
    print_header "CLEANUP"
    echo "Stopping all services..."
    # Use docker compose down to stop and remove containers, networks, and volumes.
    docker compose down -v --remove-orphans
    rm -f "$LOG_FILE" "$K6_SUMMARY_FILE" "run_test.log"
    echo "Cleanup complete."
}

# Trap EXIT signal to ensure cleanup runs
trap cleanup EXIT

# --- Main Script ---

# 0. Setup Local Tools
print_header "SETTING UP LOCAL TOOLS"
./setup_tools.sh || { echo "Failed to setup tools"; exit 1; }

# 1. Build and Start Services
print_header "BUILDING AND STARTING DOCKER SERVICES"
# Start services in detached mode and build images if they don't exist.
docker compose up --build -d || { echo "Docker Compose failed to start."; exit 1; }
echo "Services started."
# Wait a moment for services to initialize
sleep 5

# 2. Get App Server Container IDs
print_header "FETCHING CONTAINER IDs"
APP_SERVER_1_ID=$(docker compose ps -q app-server-1)
APP_SERVER_2_ID=$(docker compose ps -q app-server-2)

if [ -z "$APP_SERVER_1_ID" ] || [ -z "$APP_SERVER_2_ID" ]; then
    echo "Error: Could not retrieve container IDs for one or both app-server replicas."
    docker compose logs
    exit 1
fi
echo "Replica 1 ID: $APP_SERVER_1_ID"
echo "Replica 2 ID: $APP_SERVER_2_ID"


# 3. Build and Start Resource Monitor
print_header "BUILDING AND STARTING RESOURCE MONITOR"
(cd resource-monitor && go build -o resource-monitor .) || { echo "Failed to build resource-monitor"; exit 1; }
./resource-monitor/resource-monitor "$APP_SERVER_1_ID" "$APP_SERVER_2_ID" > "$LOG_FILE" &
MONITOR_PID=$!
echo "Resource Monitor started with PID: $MONITOR_PID, logging to $LOG_FILE"
sleep 1

# 4. Run Load Test
print_header "RUNNING k6 LOAD TEST (4000 RPS for 10s)"
(cd "$LOAD_GEN_DIR" && ../bin/k6 run k6-script.js --summary-export="../$K6_SUMMARY_FILE") > run_test.log 2>&1
K6_EXIT_CODE=$?
if [ $K6_EXIT_CODE -ne 0 ]; then
    echo "k6 load test failed with exit code $K6_EXIT_CODE."
fi
cat run_test.log | ./bin/filter_k6_output.awk

# Give the resource monitor a moment to log the final stats
sleep 2

# 5. Process and Display Results
print_header "TEST RESULTS"

# Process Resource Usage
echo "--- Resource Usage Summary (Aggregated from 2 replicas) ---"
if [ ! -f "$LOG_FILE" ] || [ ! -s "$LOG_FILE" ]; then
    echo "Log file '$LOG_FILE' not found or is empty. Cannot process resource usage."
else
    DATA_ROWS=$(tail -n +2 "$LOG_FILE")
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
    ./"$BIN_DIR"/jq -r '
        .metrics |
        {
            "HTTP Requests": .http_reqs.count,
            "RPS (actual)": .http_reqs.rate,
            "Failed Requests": .http_req_failed.value,
            "Request Duration (p95) (ms)": .http_req_duration."p(95)",
            "Request Duration (avg) (ms)": .http_req_duration.avg,
            "Browse Duration (p95) (ms)": .http_req_duration_browse."p(95)",
            "Browse Duration (avg) (ms)": .http_req_duration_browse.avg,
            "Submit Duration (p95) (ms)": .http_req_duration_submit."p(95)",
            "Submit Duration (avg) (ms)": .http_req_duration_submit.avg
        } |
        to_entries | .[] |
        if (.key | contains("Duration")) then
            "\(.key): \(.value)"
        else
            "\(.key): \(.value)"
        end
    ' "$K6_SUMMARY_FILE"
else
    echo "k6 summary file not found."
fi

echo "======================================================================"
echo "                      TEST COMPLETED"
echo "======================================================================"
