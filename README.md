# markparsr [![Go Reference](https://pkg.go.dev/badge/github.com/azyphon/markparsr.svg)](https://pkg.go.dev/github.com/azyphon/markparsr)

Markparsr ensures there's consistency between your terraform modules and markdown documentation, helping maintain documentation quality as your module evolves.

This go package analyzes both HCL files and markdown to identify gaps or inconsistencies.

## Installation

```zsh
go get github.com/azyphon/markparsr
```

## Usage

as a local test with a relative path:

```go
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
```

within github actions:

```go
func TestReadmeValidation(t *testing.T) {
	validator, err := markparsr.NewReadmeValidator(
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
```

```yaml
  - name: run global tests
    working-directory: called/tests
    run: go test -v ./...
    env:
      README_PATH: "${{ github.workspace }}/caller/README.md"
```

## Features

The markdown README is validated to contain all required sections from [terraform-docs](https://terraform-docs.io/) output, plus any additional optional content using the functional options pattern.

Automatically detects and supports both document and table output formats from terraform-docs.

It ensures all resources in your HCL Terraform code are properly documented in the README.

It checks that all resources mentioned in the README actually exist in your terraform code.

Variables and outputs are verified to match between HCL definitions and markdown documentation.

Required module files are confirmed to exist and contain content.

Urls in the markdown documentation are validated for accessibility.

## Notes

The `README_PATH` environment variable takes highest priority if set.
The path provided to NewReadmeValidator() is used if no environment variable exists.

The `MODULE_PATH` environment variable is used if set.
The directory containing the README file is used otherwise.

This approach supports both local testing and CI/CD environments with the same code.
