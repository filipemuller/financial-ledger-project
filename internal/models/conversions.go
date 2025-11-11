package models

import "math"

func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}

func FloatToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}
