/*
Markparsr ensures there's consistency between your terraform modules and markdown documentation,
helping maintain documentation quality as your module evolves. The package analyzes both
HCL files and markdown to identify gaps or inconsistencies.

key validations:
Resources/data sources match between code and docs
Variables and outputs are properly documented
Required documentation sections are present
Referenced URLs are accessible
Required files exist and aren't empty

# Standalone Usage

	package main

	import (
		"flag"
		"fmt"
		"os"

		"github.com/azyphon/markparsr"
	)

	func main() {
		readmePath := flag.String("path", "README.md", "Path to README.md file")
		flag.Parse()

		validator, err := markparsr.NewReadmeValidator(*readmePath)
		if err != nil {
			fmt.Printf("Error creating validator: %v\n", err)
			os.Exit(1)
		}

		errors := validator.Validate()
		if len(errors) > 0 {
			fmt.Printf("Found %d validation errors:\n", len(errors))
			for i, err := range errors {
				fmt.Printf("  %d. %v\n", i+1, err)
			}
			os.Exit(1)
		}

		fmt.Println("Readme validation successful!")
	}

# Test Integration

	package terraform_test

	import (
		"path/filepath"
		"testing"

		"github.com/azyphon/markparsr"
	)

	func TestDocumentation(t *testing.T) {
		readmePath := filepath.Join("..", "README.md")

		validator, err := markparsr.NewReadmeValidator(readmePath)
		if err != nil {
			t.Fatalf("Failed to create validator: %v", err)
		}

		errors := validator.Validate()

		if len(errors) > 0 {
			for _, err := range errors {
				t.Errorf("Documentation error: %v", err)
			}
		}
	}

This package validates documentation against Terraform module best practices, including:

  Standard terraform module structure with README.md
  Complete documentation of all resources, variables, and outputs
  Proper sectioning of documentation with required components

For optimal results, use markparsr alongside terraform-docs to generate parts of your documentation:

	# Generate documentation for variables and outputs
	terraform-docs markdown document --output-file README.md.new --output-mode inject .
*/
package markparsr
