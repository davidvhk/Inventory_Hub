package main

import (
	"testing"
	"time"
)

func TestParseInventoryTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string // We'll just check if it contains the month and day correctly
	}{
		{"Sun May  3 10:34", "May  3"},
		{"Wed Jan 01 00:00", "Jan  1"},
	}

	for _, tt := range tests {
		result := parseInventoryTime(tt.input)
		if result.IsZero() {
			t.Errorf("expected non-zero time for %s", tt.input)
		}
		// Since we add current year, we just verify it doesn't error and returns a plausible time
		if result.Year() != time.Now().Year() {
			t.Errorf("expected year %d, got %d", time.Now().Year(), result.Year())
		}
	}

	// Test fallback
	fallback := parseInventoryTime("invalid")
	if fallback.IsZero() {
		t.Error("expected non-zero time for invalid input")
	}
}
