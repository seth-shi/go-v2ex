package pkg

import (
	"math"
)

func BytesToMB(bytes int) float64 {
	mb := float64(bytes) / (1024 * 1024)
	return math.Round(mb*100) / 100
}
