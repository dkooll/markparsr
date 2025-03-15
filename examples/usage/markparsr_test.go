package test

import (
	"github.com/azyphon/markparsr"
	"testing"
)

// TestReadmeValidationExplicit validates Terraform module documentation.
// When running locally, this test uses the path specified in the readmePath variable.
// When running in CI/CD, environment variables README_PATH and MODULE_PATH will override
// the paths if they are set.
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
