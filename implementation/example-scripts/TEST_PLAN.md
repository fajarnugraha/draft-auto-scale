# Test Plan: Validating Application Server Resource Usage

## 1. Objective
The goal of this test is to validate the resource consumption of the `app-server` under a realistic, high-traffic load scenario. This document reflects the **current** performance characteristics of the application after several iterations of bug fixing and performance tuning.

**Target Scenario:**
- **Concurrency:** 1,000 concurrent users (VUs).
- **Throughput:** 1,000 requests per second (RPS).
- **Duration:** 10 seconds.
- **Expected Outcome:** The `app-server` should handle the full 1,000 RPS load while consuming approximately **1-2 CPU cores** and maintaining a low, stable memory footprint (under 100 MB).

## 2. Components & Implementation Strategy

The test environment consists of a load generator, a resource monitor, and a test runner script.

### A. App Server (`app-server`)
- **Memory Model:** The server now uses a **dynamic, per-request memory allocation** model. For each incoming request, it allocates a small amount of memory (e.g., 1 MB) to simulate processing data. This memory is then garbage collected, providing a realistic simulation of memory churn. This is a change from the original, static pre-allocation model.

### B. Load Generator (`k6`)
- **Location:** `doc/implementation/example-scripts/load-generator/`
- **File:** `k6-script.js`
- **Logic:**
    1.  **Configuration:** The script is configured for 1,000 virtual users and a constant arrival rate of 1,000 RPS for 10 seconds.
    2.  **User Workflow:** Each VU logs in once, then repeatedly calls the `/browse` and `/submit` endpoints in an 80/20 mix.

### C. Programmatic Resource Monitor (`resource-monitor`)
- **Location:** `doc/implementation/example-scripts/resource-monitor/`
- **Logic:**
    1.  **Input:** Takes the PID of the `app-server` as a command-line argument.
    2.  **CPU Measurement:** It reads `/proc/<PID>/stat` and calculates the CPU cores used.
        -   **Improvement:** The monitor was significantly improved. It now uses `cgo` to call the system's `sysconf(_SC_CLK_TCK)` function, ensuring the CPU time calculation is accurate and not subject to hardcoded assumptions about the system's clock tick rate.
    3.  **Memory Measurement:** It reads `/proc/<PID>/status` to get the `VmRSS` (Resident Set Size).
    4.  **Output:** It prints results in a clean, CSV format.

### D. Test Runner Script (`run_test.sh`)
- **Location:** `doc/implementation/example-scripts/`
- **Logic:** This shell script orchestrates the entire test:
    1.  **Builds** the `app-server` and `resource-monitor`.
    2.  **Starts** the `app-server` and `resource-monitor` in the background.
    3.  **Executes** the `k6` load test.
    4.  **Cleans up** all processes.
    5.  **Reports** a clear summary of resource usage and k6 performance metrics, with request durations explicitly labeled in milliseconds (ms).
