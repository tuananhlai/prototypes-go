package cheatsheettest

import (
	"errors"
	"fmt"
	"strings"
)

// ApplyDiscount calculates the new price after applying a percentage-based code.
// Format: "SAVE[Percentage]" e.g., "SAVE20"
func ApplyDiscount(basePrice float64, code string) (float64, error) {
	if !strings.HasPrefix(code, "SAVE") {
		return basePrice, errors.New("invalid code")
	}

	if basePrice <= 0 {
		return 0, errors.New("base price must be greater than 0")
	}

	// Extract the number part
	percentStr := strings.TrimPrefix(code, "SAVE")

	// Imagine we use a custom or simplified parser here
	var percent float64
	_, err := fmt.Sscanf(percentStr, "%f", &percent)
	if err != nil {
		return basePrice, errors.New("bad percentage format")
	}

	// LOGIC BUG: We forget to check if percent is greater than 100 or less than 0.
	// A fuzzer will quickly find "SAVE200" or "SAVE-500".
	discountAmount := basePrice * (percent / 100)
	finalPrice := basePrice - discountAmount

	return finalPrice, nil
}
