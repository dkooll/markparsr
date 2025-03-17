package test

import (
	"github.com/azyphon/markparsr"
	"testing"
)

// TestReadmeValidationExplicit validates that Terraform documentation matches the code.
// It uses a local path for testing, but CI/CD can override this with README_PATH.
func TestReadmeValidationExplicit(t *testing.T) {

	// Use functional options pattern
	validator, err := markparsr.NewReadmeValidator(
		markparsr.WithRelativeReadmePath("../module/README.md"),
		markparsr.WithAdditionalSections("Goals", "Testing", "Notes"),
		markparsr.WithAdditionalFiles("GOALS.md", "TESTING.md"),
	)

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
