package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCentsToFloat(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		expected float64
	}{
		{"Zero", 0, 0.0},
		{"One dollar", 100, 1.0},
		{"One hundred dollars", 10000, 100.0},
		{"Decimal amount", 10050, 100.50},
		{"Complex decimal", 12345, 123.45},
		{"Single cent", 1, 0.01},
		{"Large amount", 123456789, 1234567.89},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CentsToFloat(tt.cents)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFloatToCents(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected int64
	}{
		{"Zero", 0.0, 0},
		{"One dollar", 1.0, 100},
		{"One hundred dollars", 100.0, 10000},
		{"Decimal amount", 100.50, 10050},
		{"Complex decimal", 123.45, 12345},
		{"Single cent", 0.01, 1},
		{"Large amount", 1234567.89, 123456789},
		{"Precision handling", 100.123, 10012},
		{"Precision handling 2", 100.126, 10013},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FloatToCents(tt.amount)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that converting cents -> float -> cents maintains precision
	tests := []int64{0, 1, 100, 10050, 123456789}

	for _, cents := range tests {
		t.Run("RoundTrip", func(t *testing.T) {
			floatVal := CentsToFloat(cents)
			backToCents := FloatToCents(floatVal)
			assert.Equal(t, cents, backToCents, "Round trip conversion should maintain value")
		})
	}
}
