package models

import "math"

// CentsToFloat converts cents (int64) to float64 currency representation
// Example: 10050 cents -> 100.50
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}

// FloatToCents converts float64 currency to cents (int64)
// Rounds to nearest cent to avoid precision issues
// Example: 100.50 -> 10050 cents
func FloatToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}
