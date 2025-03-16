package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

// Validator defines the interface for all validation components.
type Validator interface {
	Validate() []error
}

// ReadmeValidator coordinates validation of Terraform module documentation.
type ReadmeValidator struct {
	readmePath string
	modulePath string
	markdown   *MarkdownContent
	terraform  *TerraformContent
	validators []Validator
}

// NewReadmeValidator creates a validator for Terraform module documentation.
// The validator needs to know where to find the README.md file.
// You can provide a path as a parameter. If you don't provide a path,
// the validator will look for the README_PATH environment variable.
// The validator will look for Terraform files in the same directory as the README.
// If you set the MODULE_PATH environment variable, it will override the Terraform files location.
func NewReadmeValidator(readmePath ...string) (*ReadmeValidator, error) {
	var finalReadmePath string

	// Determine the README path
	if len(readmePath) > 0 && readmePath[0] != "" {
		finalReadmePath = readmePath[0]
	} else {
		// Look for environment variable
		finalReadmePath = os.Getenv("README_PATH")
		if finalReadmePath == "" {
			return nil, fmt.Errorf("README path not provided and README_PATH environment variable not set")
		} else {
			fmt.Printf("using environment variable README_PATH: %s\n", finalReadmePath)
		}
	}

	absReadmePath, err := filepath.Abs(finalReadmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for README: %w", err)
	}

	// Always use the README's directory for module path
	moduleDir := filepath.Dir(absReadmePath)

	// Allow MODULE_PATH environment variable to override
	if envModulePath := os.Getenv("MODULE_PATH"); envModulePath != "" {
		moduleDir = envModulePath
	}

	absModulePath, err := filepath.Abs(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute module path: %w", err)
	}

	data, err := os.ReadFile(absReadmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	markdown := NewMarkdownContent(string(data))

	terraform, err := NewTerraformContent(absModulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terraform content: %w", err)
	}

	validator := &ReadmeValidator{
		readmePath: absReadmePath,
		modulePath: absModulePath,
		markdown:   markdown,
		terraform:  terraform,
	}

	sectionValidator := NewSectionValidator(markdown)
	validator.validators = []Validator{
		sectionValidator,
		NewFileValidator(absReadmePath, absModulePath),
		NewURLValidator(markdown),
		NewTerraformDefinitionValidator(markdown, terraform),
		NewItemValidator(markdown, terraform, "Variables", "variable", []string{"Required Inputs", "Optional Inputs"}, "variables.tf"),
		NewItemValidator(markdown, terraform, "Outputs", "output", []string{"Outputs"}, "outputs.tf"),
	}

	return validator, nil
}

// Validate runs all validators and collects their errors.
func (rv *ReadmeValidator) Validate() []error {
	var allErrors []error
	for _, validator := range rv.validators {
		allErrors = append(allErrors, validator.Validate()...)
	}
	return allErrors
}
