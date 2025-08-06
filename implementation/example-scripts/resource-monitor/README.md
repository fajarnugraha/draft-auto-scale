# Programmatic Resource Monitor

This directory contains a simple Go application that programmatically monitors the CPU and Memory usage of a given process ID (PID).

## Application: `main.go`

The monitor reads from the `/proc` filesystem to gather statistics.

-   **CPU Usage:** Calculated by reading `/proc/<PID>/stat` and measuring the change in CPU time (utime + stime) over wall-clock time. The result is expressed in "CPU Cores". For example, a value of `1.5` means the process is consuming one and a half CPU cores.
-   **Memory Usage:** Calculated by reading the `VmRSS` (Resident Set Size) field from `/proc/<PID>/status`. This represents the actual physical memory being used by the process.

### How to Run

This application is intended to be run by the main `run_test.sh` script.

To run it manually:
1.  Build the binary: `go build -o resource-monitor .`
2.  Find the PID of the process you want to monitor (e.g., `pidof app-server`).
3.  Run the monitor: `./resource-monitor <PID>`
