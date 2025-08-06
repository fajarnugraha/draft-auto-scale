# Application Server Performance Test Results

## 1. Objective
This document summarizes the results of a series of load tests performed on the `app-server`. The goal was to understand its performance characteristics and identify its scaling limits under increasingly heavy request loads.

The server was tested at four different load levels: 100, 1,000, 4,000, and 10,000 requests per second (RPS), with the number of virtual users (VUs) matching the RPS. Each test ran for 10 seconds, with the server allocating 1 MB of memory per request to simulate real-world workload.

## 2. Summary of Findings
The application scales almost perfectly linearly up to the 4,000 RPS mark, at which point performance begins to degrade significantly. The primary bottleneck under heavy load appears to be latency caused by Go's garbage collector (GC) and not raw CPU limits.

- **Excellent Performance (up to 4k RPS):** The server is highly efficient, handling up to 4,000 requests per second with minimal CPU and memory, and sub-10ms average response times.
- **Bottleneck Identified (at 10k RPS):** At the 10,000 RPS level, the server could not maintain the target throughput, achieving only ~8,000 RPS. More importantly, average latency skyrocketed from ~9ms to nearly 1,000ms (1 second), and p95 latency exceeded 3.7 seconds. This indicates a fundamental breakdown in the application's ability to handle the load within a single process.
- **Resource Scaling:** CPU and Memory usage scaled predictably with the load, but the dramatic explosion in latency is the key indicator of the performance ceiling.

## 3. Detailed Test Results

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

## 4. Conclusion
The `app-server` is a highly efficient, single-process application. Its performance is excellent for loads up to and including 4,000 requests per second.

However, the 10,000 RPS test proves that to scale beyond this point, a simple vertical scaling approach (i.e., relying on a single process on a multi-core machine) is insufficient. The latency degradation indicates that the Go runtime's garbage collector, while incredibly fast, becomes a bottleneck when faced with such an extreme rate of memory allocation and deallocation.

To achieve true linear scalability beyond 4,000 RPS, the architecture would need to evolve to a multi-process or distributed model (e.g., running multiple instances of the server behind a load balancer) to spread the GC pressure across different processes and machines.
