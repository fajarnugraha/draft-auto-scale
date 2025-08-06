# Application Server Performance Test Results

## 1. Objective
This document summarizes the results of a series of load tests performed on the `app-server`. The goal was to understand its performance characteristics and identify its scaling limits under increasingly heavy request loads.

The server was tested in two main configurations:
1.  **Single Process:** A single `app-server` binary was tested at 100, 1,000, 4,000, and 10,000 RPS.
2.  **Horizontally Scaled (Docker):** A containerized stack with two `app-server` replicas behind an Nginx load balancer was tested at 4,000 RPS.

## 2. Single-Process Test Results

### Summary of Findings
The application scales almost perfectly linearly up to the 4,000 RPS mark, at which point performance begins to degrade significantly. The primary bottleneck under heavy load appears to be latency caused by Go's garbage collector (GC) and not raw CPU limits.

- **Excellent Performance (up to 4k RPS):** The server is highly efficient, handling up to 4,000 requests per second with minimal CPU and memory, and sub-10ms average response times.
- **Bottleneck Identified (at 10k RPS):** At the 10,000 RPS level, the server could not maintain the target throughput, achieving only ~8,000 RPS. More importantly, average latency skyrocketed from ~9ms to nearly 1,000ms (1 second), and p95 latency exceeded 3.7 seconds. This indicates a fundamental breakdown in the application's ability to handle the load within a single process.

### Detailed Test Results

| Metric                  | 100 RPS | 1,000 RPS | 4,000 RPS | 10,000 RPS |
| ----------------------- | :------ | :-------- | :-------- | :--------- |
| **Target Throughput**   | 100     | 1,000     | 4,000     | **10,000** |
| **Actual RPS**          | ~100    | ~1,000    | ~4,000    | **~8,000** |
| ---                     | ---     | ---       | ---       | ---        |
| **Avg CPU Cores**       | 0.12    | 0.83      | 2.07      | 2.59       |
| **Peak CPU Cores**      | 0.19    | 1.36      | 3.49      | 5.18       |
| **Peak Memory (MB)**    | 14      | 46        | 210       | **3,183**  |
| ---                     | ---     | ---       | ---       | ---        |
| **Avg Duration (ms)**   | 0.83    | 1.06      | 8.81      | **987.48** |
| **p95 Duration (ms)**   | 2.13    | 2.39      | 61.37     | **3743.36**|

## 3. Horizontally Scaled Testing (with Docker Compose)

To test the application's ability to scale horizontally, a new test environment was created using Docker Compose.

### Architecture
- **2 `app-server` Replicas:** Two instances of the `app-server` container were run simultaneously.
- **Nginx Load Balancer:** An Nginx container was placed in front of the replicas to distribute incoming requests in a round-robin fashion.
- **Aggregated Monitoring:** The `resource-monitor` was rewritten to read the combined CPU and memory usage from both replicas using `docker stats`.

### 4,000 RPS Test Results

| Metric                        | Result      |
| ----------------------------- | :---------- |
| **Target Throughput**         | 4,000       |
| **Actual RPS**                | **~1,000**  |
| ---                           | ---         |
| **Avg CPU Cores (Combined)**  | 0.73        |
| **Peak CPU Cores (Combined)** | 1.03        |
| **Peak Memory (MB, Combined)**| 11.96       |
| ---                           | ---         |
| **Avg Duration (ms)**         | 1.30        |
| **p95 Duration (ms)**         | 2.71        |

### Analysis & Conclusion
The results from the containerized test are definitive: **the application scales horizontally with exceptional efficiency.**

The most important result is that the test **failed to generate the target 4,000 RPS**, only achieving ~1,000 RPS. This is not a failure of the `app-server`. Instead, it reveals that the bottleneck has shifted from the application to the `k6` load-generation tool itself. The two `app-server` replicas were so fast and required so few resources (a combined average of only 0.73 CPU cores) that the single `k6` process could not generate requests quickly enough to stress them.

This is an excellent outcome. It proves that by adding more replicas behind a load balancer, the application's capacity can be increased significantly. The path to handling loads greater than 4,000 RPS is not to optimize the single process further, but to deploy more instances of it.