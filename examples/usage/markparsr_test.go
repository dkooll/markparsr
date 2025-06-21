package test

import (
	"github.com/dkooll/markparsr"
	"testing"
)

func TestReadmeValidationExplicit(t *testing.T) {

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
