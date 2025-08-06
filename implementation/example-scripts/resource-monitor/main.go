package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// parseDockerStats runs `docker stats` for a given container ID and extracts CPU & Memory.
func parseDockerStats(containerID string) (float64, float64) {
	// --no-stream: Get a single snapshot.
	// --format: Specify a custom output format for easy parsing.
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", "{{.CPUPerc}},{{.MemUsage}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		// This will happen if the container stops. Return 0 for both values.
		return 0, 0
	}

	// Example output: "150.55%,12.34MiB / 1.95GiB"
	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) != 2 {
		return 0, 0
	}

	// ---
	// CPU Parsing
	// ---
	// Remove the '%' suffix and parse.
	cpuPercentStr := strings.TrimSuffix(parts[0], "%")
	cpuPercent, err := strconv.ParseFloat(cpuPercentStr, 64)
	if err != nil {
		cpuPercent = 0
	}
	// The result from docker stats is a percentage of all cores.
	// e.g., 150% on a 2-core machine means 1.5 cores are being used.
	cpuCores := cpuPercent / 100.0

	// ---
	// Memory Parsing
	// ---
	// Example: "12.34MiB / 1.95GiB" -> we only need the first part.
	memParts := strings.Fields(parts[1])
	if len(memParts) == 0 {
		return cpuCores, 0
	}
	memStr := memParts[0]
	var memMB float64
	// It could be in MiB, GiB, KiB etc. We'll parse it.
	if strings.Contains(memStr, "MiB") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "MiB"), 64)
		memMB = val
	} else if strings.Contains(memStr, "GiB") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "GiB"), 64)
		memMB = val * 1024
	} else if strings.Contains(memStr, "KiB") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(memStr, "KiB"), 64)
		memMB = val / 1024
	}

	return cpuCores, memMB
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./resource-monitor <containerID1> <containerID2> ...")
	}
	containerIDs := os.Args[1:]

	fmt.Println("Timestamp,TotalCPUCores,TotalMemoryMB")

	for {
		var totalCPU float64
		var totalMem float64
		activeContainers := 0

		for _, id := range containerIDs {
			cpu, mem := parseDockerStats(id)
			if cpu > 0 || mem > 0 {
				activeContainers++
			}
			totalCPU += cpu
			totalMem += mem
		}

		// If all containers have stopped, exit the monitor.
		if activeContainers == 0 {
			break
		}

		// ---
		// Output
		// ---
		fmt.Printf("%s,%.4f,%.2f\n", time.Now().Format(time.RFC3339), totalCPU, totalMem)

		time.Sleep(1 * time.Second)
	}
}