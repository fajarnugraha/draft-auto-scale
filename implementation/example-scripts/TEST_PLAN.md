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
        -   **Improvement:** # Test Plan: Validating Application Server Resource Usage

## 1. Objective
The goal of this test is to validate the resource consumption of the `app-server` under realistic, high-traffic load scenarios. This document outlines the components and strategies used for testing both a single-process instance and a horizontally-scaled, containerized deployment.

## 2. Test Configurations

Two primary configurations are used for testing:

### A. Single-Process Testing
- **Description:** A single `app-server` binary is run directly on the host machine.
- **Purpose:** To understand the baseline performance and identify the vertical scaling limits of a single application instance.
- **Orchestration:** The `run_test.sh` script handles the lifecycle of the `app-server` and the `resource-monitor` processes directly.

### B. Horizontally-Scaled Testing (Docker)
- **Description:** A multi-container environment managed by `docker compose`.
- **Purpose:** To validate that the application can scale horizontally and to test its performance in a more production-like, distributed environment.
- **Orchestration:** The `run_test.sh` script acts as an orchestrator for `docker compose`, starting, stopping, and monitoring the containerized services.

## 3. Core Components

### A. App Server (`app-server`)
- **Memory Model:** The server uses a **dynamic, per-request memory allocation** model. For each incoming request, it allocates a small amount of memory (e.g., 1 MB) to simulate processing data, which is then garbage collected.
- **Containerization:** A `Dockerfile` is provided to build the Go application into a minimal, production-ready Alpine Linux image.

### B. Load Generator (`k6`)
- **Location:** `doc/implementation/example-scripts/load-generator/`
- **File:** `k6-script.js`
- **Logic:** The script is configured to simulate a specific number of virtual users (VUs) and requests per second (RPS). It targets `localhost:8080` for both test configurations. In the Docker setup, this port is mapped to the Nginx load balancer.

### C. Resource Monitor (`resource-monitor`)
- **Location:** `doc/implementation/example-scripts/resource-monitor/`
- **Logic:** This tool's behavior changes based on the test configuration:
    1.  **Single-Process Mode:** It takes a Process ID (PID) as an argument and reads from the `/proc` filesystem to measure resource usage.
    2.  **Docker Mode:** It is rewritten to take Docker container IDs as arguments. It then uses the `docker stats` command to get CPU and memory data, which it aggregates from all specified containers before printing a unified CSV output.

### D. Nginx Load Balancer (`nginx/`)
- **Purpose:** Used exclusively in the Docker-based test configuration.
- **Logic:** It performs simple round-robin load balancing, distributing incoming traffic from the `k6` client across the two `app-server` container replicas.

### E. Test Runner Script (`run_test.sh`)
- **Location:** `doc/implementation/example-scripts/`
- **Logic:** This is the primary entry point for all tests. It now contains logic to detect the desired test mode and orchestrate either the single-process run or the full `docker compose` lifecycle.

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
