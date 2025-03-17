package test

import (
	"github.com/azyphon/markparsr"
	"testing"
)

// TestReadmeValidationExplicit validates that Terraform documentation matches the code.
// It uses a local path for testing, but CI/CD can override this with README_PATH.
func TestReadmeValidationExplicit(t *testing.T) {
	readmePath := "../module/README.md"

	// Create custom options with specific additional sections to validate
	options := markparsr.DefaultOptions()

	options.AdditionalSections = []string{"Goals", "Testing", "Notes"}
	options.AdditionalFiles = []string{"GOALS.md", "TESTING.md"}

	// Use options with autodetect format and additional sections
	validator, err := markparsr.NewReadmeValidatorWithOptions(options, readmePath)
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
