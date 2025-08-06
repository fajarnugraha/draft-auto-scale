package main

/*
#include <unistd.h>
*/
import "C"
import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// getClockTicks returns the number of clock ticks per second.
func getClockTicks() float64 {
	// sysconf(_SC_CLK_TCK) is the standard C library way to get this value.
	return float64(C.sysconf(C._SC_CLK_TCK))
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./resource-monitor <PID>")
	}
	pid := os.Args[1]

	// Get system clock ticks per second (jiffies)
	ticksPerSecond := getClockTicks()
	if ticksPerSecond <= 0 {
		log.Println("Warning: Could not determine system clock ticks. Defaulting to 100.")
		ticksPerSecond = 100 // Fallback to a common default
	}

	var lastTotalCPUTime float64
	lastSampleTime := time.Now()

	fmt.Println("Timestamp,CPUCores,MemoryMB")

	for {
		// --- Memory Measurement ---
		memPath := fmt.Sprintf("/proc/%s/status", pid)
		memBytes, err := ioutil.ReadFile(memPath)
		if err != nil {
			// Process likely terminated, exit gracefully
			break
		}
		memLines := strings.Split(string(memBytes), "\n")
		var rssMB float64
		for _, line := range memLines {
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					rssKB, _ := strconv.ParseFloat(fields[1], 64)
					rssMB = rssKB / 1024
				}
				break
			}
		}

		// --- CPU Measurement ---
		statPath := fmt.Sprintf("/proc/%s/stat", pid)
		statBytes, err := ioutil.ReadFile(statPath)
		if err != nil {
			break // Process likely terminated
		}
		statFields := strings.Fields(string(statBytes))
		if len(statFields) < 15 {
			// stat file is not as expected, skip this sample
			continue
		}
		utime, _ := strconv.ParseFloat(statFields[13], 64)
		stime, _ := strconv.ParseFloat(statFields[14], 64)
		totalCPUTime := utime + stime

		// --- Calculation ---
		currentTime := time.Now()
		elapsedWallTime := currentTime.Sub(lastSampleTime).Seconds()
		elapsedCPUTime := (totalCPUTime - lastTotalCPUTime) / ticksPerSecond

		cpuCores := 0.0
		if elapsedWallTime > 0 {
			cpuCores = elapsedCPUTime / elapsedWallTime
		}

		// --- Output ---
		fmt.Printf("%s,%.4f,%.2f\n", currentTime.Format(time.RFC3339), cpuCores, rssMB)

		// --- Update for next iteration ---
		lastTotalCPUTime = totalCPUTime
		lastSampleTime = currentTime

		time.Sleep(1 * time.Second)
	}
}
