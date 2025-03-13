package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

type Validator interface {
	Validate() []error
}

type ReadmeValidator struct {
	readmePath string
	markdown   *MarkdownContent
	terraform  *TerraformContent
	validators []Validator
}

func NewReadmeValidator(readmePath string) (*ReadmeValidator, error) {
	if envPath := os.Getenv("README_PATH"); envPath != "" {
		readmePath = envPath
	}

	absReadmePath, err := filepath.Abs(readmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absReadmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	markdown := NewMarkdownContent(string(data))

	terraform, err := NewTerraformContent()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terraform content: %w", err)
	}

	validator := &ReadmeValidator{
		readmePath: absReadmePath,
		markdown:   markdown,
		terraform:  terraform,
	}

	sectionValidator := NewSectionValidator(markdown)
	validator.validators = []Validator{
		sectionValidator,
		NewFileValidator(absReadmePath),
		NewURLValidator(markdown),
		NewTerraformDefinitionValidator(markdown, terraform),
		NewItemValidator(markdown, terraform, "Variables", "variable", []string{"Required Inputs", "Optional Inputs"}, "variables.tf"),
		NewItemValidator(markdown, terraform, "Outputs", "output", []string{"Outputs"}, "outputs.tf"),
	}

	return validator, nil
}

func (rv *ReadmeValidator) Validate() []error {
	var allErrors []error
	for _, validator := range rv.validators {
		allErrors = append(allErrors, validator.Validate()...)
	}
	return allErrors
}
