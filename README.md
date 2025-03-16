# markparsr [![Go Reference](https://pkg.go.dev/badge/github.com/azyphon/markparsr.svg)](https://pkg.go.dev/github.com/azyphon/markparsr)

Markparsr ensures there's consistency between your terraform modules and markdown documentation, helping maintain documentation quality as your module evolves.

This go package analyzes both HCL files and markdown to identify gaps or inconsistencies.

## Installation

```bash
go get github.com/azyphon/markparsr
```

## Usage

```go
func TestTerraformDocumentation(t *testing.T) {
    // For local testing with a relative path
    validator, err := markparsr.NewReadmeValidator("../README.md")
    if err != nil {
        t.Fatalf("Failed to create validator: %v", err)
    }

    errors := validator.Validate()
    if len(errors) > 0 {
        t.Errorf("Found documentation errors:")
        for _, err := range errors {
            t.Errorf("  - %v", err)
        }
    }
}
```

```yaml
# GitHub Actions example
- name: Validate Documentation
  run: go test -v ./tests
  env:
    README_PATH: "${{ github.workspace }}/README.md"
```

## Features

The markdown README is validated to contain all required sections from terraform-docs output, plus any additional optional content.

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
