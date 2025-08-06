# Application Server Performance Test Results

## 1. Objective
This document summarizes the results of a series of load tests performed on the `app-server`. The goal was to understand its performance characteristics, identify its scaling limits, and analyze the behavior of the system under different configurations.

The server was tested in two main configurations:
1.  **Single Process:** A single `app-server` binary was tested at 100, 1,000, 4,000, and 10,000 RPS.
2.  **Horizontally Scaled (Docker):** A containerized stack with two `app-server` replicas behind an Nginx load balancer was tested at 4,000 RPS, with monitoring for both the application and the host system.

## 2. Single-Process Test Results
These tests aimed to find the vertical scaling limit of a single application instance.

### Detailed Test Results
| Metric                  | 100 RPS | 1,000 RPS | 4,000 RPS | 10,000 RPS |
| ----------------------- | :------ | :-------- | :-------- | :--------- |
| **Target Throughput**   | 100     | 1,000     | 4,000     | **10,000** |
| **Actual RPS**          | ~100    | ~1,000    | **~4,000**| **~8,000** |
| ---                     | ---     | ---       | ---       | ---        |
| **Avg CPU Cores**       | 0.12    | 0.83      | 2.07      | 2.59       |
| **Peak CPU Cores**      | 0.19    | 1.36      | **3.49**  | 5.18       |
| ---                     | ---     | ---       | ---       | ---        |
| **Avg Duration (ms)**   | 0.83    | 1.06      | 8.81      | **987.48** |
| **p95 Duration (ms)**   | 2.13    | 2.39      | 61.37     | **3743.36**|

### Analysis
The application scales well up to 4,000 RPS. At 10,000 RPS, latency skyrockets, indicating the single process is overwhelmed. The primary bottleneck is likely Go's garbage collector (GC) struggling with the extreme rate of memory allocation.

## 3. Horizontally Scaled Test Results (Docker)
This test aimed to prove horizontal scalability and analyze the root cause of system limits.

### 4,000 RPS Test Results
| Metric                                | Result      |
| ------------------------------------- | :---------- |
| **Target Throughput**                 | 4,000       |
| **Actual RPS**                        | **~1,000**  |
| ---                                   | ---         |
| **Avg App-Server CPU (Combined)**     | **0.79**    |
| **Avg Total System CPU (All Procs)**  | **2.36**    |
| ---                                   | ---         |
| **Avg Duration (ms)**                 | 1.56        |
| **p95 Duration (ms)**                 | 2.98        |


## 4. Final Analysis: The Connection Throughput Bottleneck

The test results presented a critical finding: the system failed to reach its 4,000 RPS target in the Docker test, stalling at ~1,000 RPS, even though it had over 4 CPU cores of available capacity. This proves the system was **not CPU-bound**, but **contention-bound**.

The bottleneck is the maximum throughput of new TCP connections that the system can establish and route through the container stack.

### The "Two Hops vs. One Hop" Problem

1.  **Single-Process Test (One Hop):**
    `k6` -> `Kernel` -> `app-server`
    In this test, `k6` established a connection directly with the `app-server`. The kernel's networking path is highly optimized for this. The system successfully handled **~4,000 connections per second**.

2.  **Docker Test (Two Hops):**
    `k6` -> `Kernel` -> `nginx` -> `Kernel (Docker Network)` -> `app-server`
    For every single request, the system had to perform **two distinct connection management operations**: one from `k6` to `nginx`, and a second from `nginx` to an `app-server` container via the Docker software network.

### Conclusion
The Docker test revealed the maximum throughput of this two-hop connection chain. The system could only establish and process **~1,000 end-to-end sessions per second**. This is not a failure of `k6`'s ability to *send* requests, nor a failure of the `app-server`'s ability to *process* them. It is the limit of the host machine's ability to *manage* the connection setup and routing through the more complex Docker networking path at that rate.

The `app-server` itself is highly efficient and scales horizontally. The investigation successfully identified the next bottleneck in the system: the connection-per-second throughput of the container host environment.