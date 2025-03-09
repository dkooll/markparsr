package main

import (
	"fmt"
	"os"

	"github.com/dkooll/markparsr"
)

func main() {
	// Example 1: Using default configuration
	fmt.Println("Validating README with default configuration...")
	validator, err := markparsr.New(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating validator: %v\n", err)
		os.Exit(1)
	}

	errors := validator.Validate()
	if len(errors) > 0 {
		fmt.Printf("Found %d validation errors:\n", len(errors))
		for i, err := range errors {
			fmt.Printf("%d. %v\n", i+1, err)
		}
		fmt.Println("Validation failed!")
		os.Exit(1)
	}

	fmt.Println("Validation successful!")

	// Example 2: Using custom configuration
	fmt.Println("\nValidating README with custom configuration...")
	customConfig := &markparsr.Config{
		ReadmePath:              "README.md",
		SkipURLValidation:       true,
		SkipTerraformValidation: true,
	}

	customValidator, err := markparsr.New(customConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating validator: %v\n", err)
		os.Exit(1)
	}

	customErrors := customValidator.Validate()
	if len(customErrors) > 0 {
		fmt.Printf("Found %d validation errors with custom config:\n", len(customErrors))
		for i, err := range customErrors {
			fmt.Printf("%d. %v\n", i+1, err)
		}
		fmt.Println("Validation with custom config failed!")
		os.Exit(1)
	}

	fmt.Println("Validation with custom config successful!")
}
