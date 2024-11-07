package timer

import (
	"time"
)

// EstimateMinClockResolution estimates the system's minimum clock resolution
func EstimateMinClockResolution(iterations int) time.Duration {
	var minDiff time.Duration = time.Minute // Initialize with a large duration

	for i := 0; i < iterations; i++ {
		t1 := time.Now()
		t2 := time.Now()
		diff := t2.Sub(t1) // https://pkg.go.dev/time@master#hdr-Monotonic_Clocks
		if diff > 0 && diff < minDiff {
			minDiff = diff
		}
	}

	return minDiff
}
