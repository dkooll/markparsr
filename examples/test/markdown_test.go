package test

import (
	"testing"

	"github.com/azyphon/markparsr"
)

func TestReadmeValidation(t *testing.T) {
	// You can specify a custom path or use the default "README.md"
	validator, err := markparsr.NewReadmeValidator("../../README.md")
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
