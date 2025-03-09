package test

import (
	"testing"

	"github.com/azyphon/markparsr"
)

// TestReadmeValidation demonstrates how to use markparsr in a test
func TestReadmeValidation(t *testing.T) {
	// Create a validator with custom configuration
	config := &markparsr.Config{
		ReadmePath:         "../README.md", // Path relative to this test file
		SkipURLValidation: true,           // Skip URL validation to avoid network requests in tests
	}

	validator, err := markparsr.New(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Run validation
	errors := validator.Validate()

	// Check for errors
	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("Validation error: %v", err)
		}
	}
}

// TestWithDefaultConfig demonstrates using the default configuration
func TestWithDefaultConfig(t *testing.T) {
	// When passing nil, the default config is used
	validator, err := markparsr.New(nil)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Run validation with fail-fast
	err = validator.ValidateWithFailFast()
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}
