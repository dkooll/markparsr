package test

import (
	"testing"

	"github.com/azyphon/markparsr"
)

func TestReadmeValidation(t *testing.T) {
	config := &markparsr.Config{
		ReadmePath:         "../README.md",
		SkipURLValidation: true,
	}

	validator, err := markparsr.New(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	errors := validator.Validate()

	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("Validation error: %v", err)
		}
	}
}

// default configuration
func TestWithDefaultConfig(t *testing.T) {
	validator, err := markparsr.New(nil)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	err = validator.ValidateWithFailFast()
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

// heading based
func TestReadmeHeaderValidation(t *testing.T) {
    config := &markparsr.Config{
        ReadmePath:      "../README.md",
        SkipURLValidation: true,
        ValidationStyle: "heading", // Use heading-based validation instead of tables
    }
    validator, err := markparsr.New(config)
    if err != nil {
        t.Fatalf("Failed to create validator: %v", err)
    }
    errors := validator.Validate()
    if len(errors) > 0 {
        for _, err := range errors {
            t.Errorf("Validation error: %v", err)
        }
    }
}
