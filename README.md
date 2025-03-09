# markparsr

A Go package for validating markdown documentation for Terraform modules, ensuring that it is complete, accurate, and follows best practices.

## Features

- Validate section headers and table structure
- Verify required files exist and are not empty
- Check URLs for accessibility
- Validate Terraform resources match those documented in markdown
- Ensure Terraform variables and outputs are properly documented

## Installation

```bash
go get github.com/yourusername/markparsr
```

## Usage

markparsr can be used in two main ways: as part of your tests or in regular code.

### Option 1: In your tests

```go
import (
    "testing"

    "github.com/yourusername/markparsr"
)

func TestReadmeValidation(t *testing.T) {
    // Create a validator with custom configuration
    config := &markparsr.Config{
        ReadmePath:         "./README.md",
        SkipURLValidation: true, // Skip URL validation in tests
    }

    validator, err := markparsr.New(config)
    if err != nil {
        t.Fatalf("Failed to create validator: %v", err)
    }

    // Run validation
    errors := validator.Validate()

    // Check for errors
    if len(errors) > 0 {
        for _, err := range errors {
            t.Errorf("Validation error: %v", err)
        }
    }
}
```

### Option 2: In regular code

```go
import (
    "fmt"
    "os"

    "github.com/yourusername/markparsr"
)

func main() {
    // Create a validator with default configuration
    validator, err := markparsr.New(nil)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating validator: %v\n", err)
        os.Exit(1)
    }

    // Run validation
    errors := validator.Validate()
    if len(errors) > 0 {
        fmt.Printf("Found %d validation errors:\n", len(errors))
        for i, err := range errors {
            fmt.Printf("%d. %v\n", i+1, err)
        }
        os.Exit(1)
    }

    fmt.Println("Validation successful!")
}
```

## Configuration

markparsr can be configured through the `Config` struct:

```go
type Config struct {
    // Path to the README.md file
    ReadmePath string

    // Skip validation of URLs in the markdown
    SkipURLValidation bool

    // Skip validation of required files
    SkipFileValidation bool

    // Skip validation of Terraform definitions
    SkipTerraformValidation bool

    // Skip validation of Terraform variables
    SkipVariablesValidation bool

    // Skip validation of Terraform outputs
    SkipOutputsValidation bool
}
```

## Validators

markparsr includes several validators:

1. **SectionValidator**: Ensures required sections are present with proper structure
2. **FileValidator**: Verifies required files exist and are not empty
3. **URLValidator**: Checks if URLs in the markdown are accessible
4. **TerraformDefinitionValidator**: Validates Terraform resources match those in markdown
5. **ItemValidator**: Ensures Terraform variables and outputs are properly documented

## License

This project is licensed under the MIT License - see the LICENSE file for details.
