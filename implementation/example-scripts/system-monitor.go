package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// cpuStats holds the raw values from /proc/stat
type cpuStats struct {
	user, nice, system, idle, iowait, irq, softirq, steal, guest, guestNice float64
}

// total returns the total non-idle CPU time.
func (s cpuStats) total() float64 {
	return s.user + s.nice + s.system + s.irq + s.softirq + s.steal + s.guest + s.guestNice
}

// totalIdle returns the total idle CPU time.
func (s cpuStats) totalIdle() float64 {
	return s.idle + s.iowait
}

// readSystemCPUStats reads the first line of /proc/stat and parses it.
func readSystemCPUStats() (cpuStats, error) {
	var stats cpuStats
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return stats, err
	}

	lines := strings.Split(string(contents), "\n")
	if len(lines) == 0 {
		return stats, fmt.Errorf("empty /proc/stat file")
	}

	// The first line is the aggregate for all CPUs.
	fields := strings.Fields(lines[0])
	if len(fields) < 8 || fields[0] != "cpu" {
		return stats, fmt.Errorf("unexpected format in /proc/stat")
	}

	stats.user, _ = strconv.ParseFloat(fields[1], 64)
	stats.nice, _ = strconv.ParseFloat(fields[2], 64)
	stats.system, _ = strconv.ParseFloat(fields[3], 64)
	stats.idle, _ = strconv.ParseFloat(fields[4], 64)
	stats.iowait, _ = strconv.ParseFloat(fields[5], 64)
	stats.irq, _ = strconv.ParseFloat(fields[6], 64)
	stats.softirq, _ = strconv.ParseFloat(fields[7], 64)
	// These fields might not exist on older kernels.
	if len(fields) > 8 {
		stats.steal, _ = strconv.ParseFloat(fields[8], 64)
	}
	if len(fields) > 9 {
		stats.guest, _ = strconv.ParseFloat(fields[9], 64)
	}
	if len(fields) > 10 {
		stats.guestNice, _ = strconv.ParseFloat(fields[10], 64)
	}

	return stats, nil
}

func main() {
	numCPU := float64(runtime.NumCPU())
	lastStats, err := readSystemCPUStats()
	if err != nil {
		log.Fatalf("Could not read initial system CPU stats: %v", err)
	}

	fmt.Println("Timestamp,TotalSystemCoresUsed")

	for {
		time.Sleep(1 * time.Second)

		currentStats, err := readSystemCPUStats()
		if err != nil {
			// Stop gracefully if we can no longer read the file.
			break
		}

		lastTotal := lastStats.total() + lastStats.totalIdle()
		currentTotal := currentStats.total() + currentStats.totalIdle()

		totalDiff := currentTotal - lastTotal
		idleDiff := currentStats.totalIdle() - lastStats.totalIdle()

		cpuUsagePercentage := 0.0
		if totalDiff > 0 {
			cpuUsagePercentage = (totalDiff - idleDiff) / totalDiff * 100
		}

		// Convert the overall percentage to the number of cores being used.
		coresUsed := cpuUsagePercentage / 100.0 * numCPU

		fmt.Printf("%s,%.4f\n", time.Now().Format(time.RFC3339), coresUsed)

		lastStats = currentStats
	}
}
