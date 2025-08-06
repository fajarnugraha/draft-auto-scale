# Test Plan: Validating Application Server Resource Usage

## 1. Objective
The goal of this test is to validate the assumptions made about the resource consumption of the `app-server`. We will measure its CPU and Memory usage under a specific, high-traffic load scenario to see if the default load simulation values are reasonable.

**Target Scenario:**
- **Concurrency:** 1,000 concurrent users (VUs).
- **Throughput:** 1,000 requests per second (RPS).
- **Duration:** 60 seconds.
- **Expected Outcome:** The `app-server` should consume approximately **2 full CPU cores** and **1 GB of memory**.

## 2. Components & Implementation Strategy

I will create three new components to achieve this: a load generator, a resource monitor, and a test runner script.

### A. Load Generator (`k6`)
- **Location:** `doc/implementation/example-scripts/load-generator/`
- **Tool:** I will use `k6`, as it is already referenced in the project's documentation (`doc/implementation/load-generator/k6.md`) and is ideal for this kind of scripted load testing.
- **File:** `k6-script.js`
- **Logic:**
    1.  **Configuration:** The script will be configured for 1,000 virtual users and a constant arrival rate of 1,000 RPS for 60 seconds.
    2.  **User Workflow:** Each virtual user will execute a realistic workflow:
        -   Log in once at the beginning of their session to get an auth token.
        -   Repeatedly call the `/browse` and `/submit` endpoints for the duration of the test.
    3.  **API Mix:** To simulate a "reasonable" traffic pattern, the requests will be weighted. A good starting point is:
        -   **80% `/browse` requests:** The most common action.
        -   **20% `/submit` requests:** A less frequent but important action.
        -   The `/login` call happens once per user and does not count toward the main RPS mix.

### B. Programmatic Resource Monitor
- **Location:** `doc/implementation/example-scripts/resource-monitor/`
- **Tool:** I will write a simple, standalone Go application. This avoids external dependencies and allows for precise, programmatic measurement by reading from the `/proc` filesystem, as requested.
- **Files:** `main.go`, `go.mod`
- **Logic:**
    1.  **Input:** The monitor will take the Process ID (PID) of the `app-server` as a command-line argument.
    2.  **Execution:** It will run in a loop, sampling data every 1 second.
    3.  **CPU Measurement:** It will read `/proc/<PID>/stat` to get the `utime` and `stime` (CPU time in jiffies). By calculating the change in CPU time versus wall-clock time, it can accurately determine the number of CPU cores being used.
    4.  **Memory Measurement:** It will read `/proc/<PID>/status` to get the `VmRSS` (Resident Set Size), which is a good measure of the actual physical memory the process is using.
    5.  **Output:** It will print the results to standard output in a simple, clean format (e.g., `Timestamp, CPUCores, MemoryMB`).

### C. Test Runner Script
- **Location:** `doc/implementation/example-scripts/`
- **File:** `run_test.sh`
- **Logic:** This shell script will orchestrate the entire test process:
    1.  **Build:** It will first build the `app-server` and `resource-monitor` Go applications.
    2.  **Start Server:** It will start the `app-server` in the background and capture its PID.
    3.  **Start Monitor:** It will start the `resource-monitor` in the background, passing the server's PID to it and redirecting its output to a `results.csv` file.
    4.  **Run Load Test:** It will execute the `k6 run` command and wait for it to complete.
    5.  **Cleanup:** Once the `k6` test is finished, it will terminate the `app-server` and `resource-monitor` processes.
    6.  **Report:** It will print a summary of the average and peak resource usage from `results.csv`.

## 3. Proposed File Structure

```
/data/telkom/auto-scale/doc/implementation/example-scripts/
├── app-server/
│   ├── ... (existing files)
├── load-generator/
│   ├── k6-script.js
│   └── README.md
├── resource-monitor/
│   ├── main.go
│   ├── go.mod
│   └── README.md
├── run_test.sh
└── TEST_PLAN.md
```

## 4. Approval
Please review this plan. If you approve, I will proceed with creating these files and writing the necessary code.
