// Package markparsr provides utilities for validating markdown documentation
// for Terraform modules, specifically focusing on README files and ensuring they
// match the actual Terraform code.
package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

// Validator is an interface for all validators
type Validator interface {
	Validate() []error
}

// MarkdownValidator orchestrates all validations
type MarkdownValidator struct {
	ReadmePath string
	Data       string
	validators []Validator
}

// Config holds configuration options for the validator
type Config struct {
	// ReadmePath is the path to the README.md file
	ReadmePath string

	// SkipURLValidation skips validation of URLs in the markdown
	SkipURLValidation bool

	// SkipFileValidation skips validation of required files
	SkipFileValidation bool

	// SkipTerraformValidation skips validation of Terraform definitions
	SkipTerraformValidation bool

	// SkipVariablesValidation skips validation of Terraform variables
	SkipVariablesValidation bool

	// SkipOutputsValidation skips validation of Terraform outputs
	SkipOutputsValidation bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		ReadmePath: "README.md",
	}
}

// New creates a new MarkdownValidator with the given configuration
func New(config *Config) (*MarkdownValidator, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Allow overriding the README path via environment variable
	readmePath := config.ReadmePath
	if envPath := os.Getenv("README_PATH"); envPath != "" {
		readmePath = envPath
	}

	absReadmePath, err := filepath.Abs(readmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	dataBytes, err := os.ReadFile(absReadmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	data := string(dataBytes)

	mv := &MarkdownValidator{
		ReadmePath: absReadmePath,
		Data:       data,
	}

	// Initialize validators based on configuration
	mv.validators = []Validator{
		NewSectionValidator(data),
	}

	if !config.SkipFileValidation {
		mv.validators = append(mv.validators, NewFileValidator(absReadmePath))
	}

	if !config.SkipURLValidation {
		mv.validators = append(mv.validators, NewURLValidator(data))
	}

	if !config.SkipTerraformValidation {
		mv.validators = append(mv.validators, NewTerraformDefinitionValidator(data))
	}

	if !config.SkipVariablesValidation {
		mv.validators = append(mv.validators, NewItemValidator(data, "Variables", "variable", "Inputs", "variables.tf"))
	}

	if !config.SkipOutputsValidation {
		mv.validators = append(mv.validators, NewItemValidator(data, "Outputs", "output", "Outputs", "outputs.tf"))
	}

	return mv, nil
}

// Validate runs all registered validators
func (mv *MarkdownValidator) Validate() []error {
	var allErrors []error
	for _, validator := range mv.validators {
		allErrors = append(allErrors, validator.Validate()...)
	}
	return allErrors
}

// ValidateWithFailFast runs validators and returns on first error
func (mv *MarkdownValidator) ValidateWithFailFast() error {
	for _, validator := range mv.validators {
		errors := validator.Validate()
		if len(errors) > 0 {
			return errors[0]
		}
	}
	return nil
}
