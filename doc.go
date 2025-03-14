/*
Package markparsr provides tools for validating Terraform module documentation.

markparsr ensures that markdown documentation for Terraform modules is complete,
accurate, and follows best practices. It validates that all resources, data sources,
inputs, and outputs mentioned in Terraform code are properly documented, and vice versa.

# Core Features

  - Validate presence of required documentation sections
  - Check that all Terraform resources and data sources are documented
  - Verify that all inputs (variables) are documented
  - Ensure that all outputs are documented
  - Validate that referenced URLs are accessible
  - Check for existence of required files

# Basic Usage

Create a validator and run validation:

	validator, err := markparsr.NewReadmeValidator("path/to/README.md")
	if err != nil {
		// Handle error
	}

	errors := validator.Validate()
	if len(errors) > 0 {
		// Handle validation errors
	}

# Validation Components

The package provides several validators that can be used individually or together:

  - SectionValidator: Ensures required sections exist in the documentation
  - FileValidator: Checks required files exist and are not empty
  - URLValidator: Validates URLs in the documentation are accessible
  - TerraformDefinitionValidator: Compares Terraform resources/data sources with documentation
  - ItemValidator: Validates variables and outputs are properly documented

# Environment Variables

The package recognizes these environment variables:

  - README_PATH: Override path to README.md
  - GITHUB_WORKSPACE: Base directory for Terraform file scanning (defaults to current directory)
*/
package markparsr
