package test

import (
	"github.com/azyphon/markparsr"
	"testing"
)

// TestReadmeValidationExplicit validates that Terraform documentation matches the code.
// It uses a local path for testing, but CI/CD can override this with README_PATH.
func TestReadmeValidationExplicit(t *testing.T) {
	readmePath := "../module/README.md"

	validator, err := markparsr.NewReadmeValidator(readmePath)
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
