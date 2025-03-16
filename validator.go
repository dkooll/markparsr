package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Validator defines the interface for all validation components
type Validator interface {
	Validate() []error
}

// Options configures the behavior of the markdown validator
type Options struct {
	// Format specifies whether the terraform-docs output is in document or table format
	Format MarkdownFormat
}

// DefaultOptions provides sensible defaults for the validator
func DefaultOptions() Options {
	return Options{
		Format: FormatAuto,
	}
}

// ReadmeValidator coordinates validation of Terraform module documentation
type ReadmeValidator struct {
	readmePath string
	modulePath string
	markdown   *MarkdownContent
	terraform  *TerraformContent
	validators []Validator
	options    Options
}

// NewReadmeValidator creates a validator with auto-format detection
func NewReadmeValidator(readmePath ...string) (*ReadmeValidator, error) {
	return NewReadmeValidatorWithOptions(DefaultOptions(), readmePath...)
}

// handleFormatEnvironment checks for format override in environment variables
func handleFormatEnvironment(options *Options) {
	if envFormat := os.Getenv("FORMAT"); envFormat != "" {
		switch strings.ToLower(envFormat) {
		case "document":
			options.Format = FormatDocument
		case "table":
			options.Format = FormatTable
		case "auto":
			options.Format = FormatAuto
		default:
			fmt.Printf("Unknown format in FORMAT environment variable: %s, using auto-detection\n", envFormat)
		}
	}
}

// NewReadmeValidatorWithOptions creates a validator with custom options
func NewReadmeValidatorWithOptions(options Options, readmePath ...string) (*ReadmeValidator, error) {
	// Determine the README path from args or environment
	readmeFile, err := resolveReadmePath(readmePath...)
	if err != nil {
		return nil, err
	}

	// Get module path from README directory or environment
	modulePath := filepath.Dir(readmeFile)
	if envModulePath := os.Getenv("MODULE_PATH"); envModulePath != "" {
		modulePath = envModulePath
	}

	absModulePath, err := filepath.Abs(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute module path: %w", err)
	}

	// Read the README content
	data, err := os.ReadFile(readmeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check for format override in environment
	handleFormatEnvironment(&options)

	// Initialize content analyzers
	markdown := NewMarkdownContent(string(data), options.Format)

	terraform, err := NewTerraformContent(absModulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terraform content: %w", err)
	}

	// Create validator with all components
	validator := &ReadmeValidator{
		readmePath: readmeFile,
		modulePath: absModulePath,
		markdown:   markdown,
		terraform:  terraform,
		options:    options,
	}

	// Initialize all validators
	validator.validators = []Validator{
		NewSectionValidator(markdown),
		NewFileValidator(readmeFile, absModulePath),
		NewURLValidator(markdown),
		NewTerraformDefinitionValidator(markdown, terraform),
		NewItemValidator(markdown, terraform, "Variables", "variable", []string{"Required Inputs", "Optional Inputs"}, "variables.tf"),
		NewItemValidator(markdown, terraform, "Outputs", "output", []string{"Outputs"}, "outputs.tf"),
	}

	return validator, nil
}

// Validate runs all validators and collects their errors.
func (rv *ReadmeValidator) Validate() []error {
	// We already printed the format when it was detected, no need to repeat it here

	var allErrors []error

	// Run all validators and collect errors
	for _, validator := range rv.validators {
		validatorErrors := validator.Validate()
		allErrors = append(allErrors, validatorErrors...)
	}

	return allErrors
}

// GetFormat returns the detected markdown format.
// This can be useful for debugging or reporting the detected format.
func (rv *ReadmeValidator) GetFormat() MarkdownFormat {
	if rv.markdown != nil {
		return rv.markdown.format
	}
	return FormatAuto
}

// resolveReadmePath gets the README path from parameters or environment
func resolveReadmePath(readmePath ...string) (string, error) {
	var finalReadmePath string

	// Try to get path from function args
	if len(readmePath) > 0 && readmePath[0] != "" {
		finalReadmePath = readmePath[0]
	} else {
		// Try to get from environment
		finalReadmePath = os.Getenv("README_PATH")
		if finalReadmePath == "" {
			return "", fmt.Errorf("README path not provided and README_PATH environment variable not set")
		}
		if os.Getenv("VERBOSE") == "true" {
			fmt.Printf("Using README_PATH from environment: %s\n", finalReadmePath)
		}
	}

	// Convert to absolute path
	absReadmePath, err := filepath.Abs(finalReadmePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for README: %w", err)
	}

	return absReadmePath, nil
}
