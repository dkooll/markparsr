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
	ReadmePath    string
	Data          string
	validators    []Validator
	foundSections map[string]bool // Used by heading-based validation
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

	// ValidationStyle determines which validation approach to use
	// "table" - expects markdown with tables (default)
	// "heading" - expects markdown with hierarchical headings
	ValidationStyle string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		ReadmePath:     "README.md",
		ValidationStyle: "table", // Default to table-based validation
	}
}

// New creates a new MarkdownValidator with the given configuration
func New(config *Config) (*MarkdownValidator, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Default validation style is table-based if not specified
	if config.ValidationStyle == "" {
		config.ValidationStyle = "table"
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
		foundSections: make(map[string]bool),
	}

	// Set up validators based on validation style
	if config.ValidationStyle == "heading" {
		// Heading-based validation
		sectionValidator := NewHeadingSectionValidator(data)
		mv.validators = []Validator{
			sectionValidator,
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
			mv.validators = append(mv.validators, NewHeadingItemValidator(data, "Variables", "variable", []string{"Required Inputs", "Optional Inputs"}, "variables.tf"))
		}

		if !config.SkipOutputsValidation {
			mv.validators = append(mv.validators, NewHeadingItemValidator(data, "Outputs", "output", []string{"Outputs"}, "outputs.tf"))
		}

		// Store the found sections for later use
		for _, section := range sectionValidator.sections {
			mv.foundSections[section] = sectionValidator.validateSection(section)
		}
	} else {
		// Table-based validation (default)
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
	}

	return mv, nil
}

// Validate runs all registered validators
func (mv *MarkdownValidator) Validate() []error {
	var allErrors []error
	for _, validator := range mv.validators {
		// For heading-style validation, check if the section exists
		if headingItemValidator, ok := validator.(*HeadingItemValidator); ok {
			// Check if any of the sections exist
			sectionExists := false
			for _, section := range headingItemValidator.sections {
				if mv.foundSections[section] {
					sectionExists = true
					break
				}
			}
			if !sectionExists {
				continue
			}
		}

		allErrors = append(allErrors, validator.Validate()...)
	}
	return allErrors
}

// ValidateWithFailFast runs validators and returns on first error
func (mv *MarkdownValidator) ValidateWithFailFast() error {
	for _, validator := range mv.validators {
		// For heading-style validation, check if the section exists
		if headingItemValidator, ok := validator.(*HeadingItemValidator); ok {
			// Check if any of the sections exist
			sectionExists := false
			for _, section := range headingItemValidator.sections {
				if mv.foundSections[section] {
					sectionExists = true
					break
				}
			}
			if !sectionExists {
				continue
			}
		}

		errors := validator.Validate()
		if len(errors) > 0 {
			return errors[0]
		}
	}
	return nil
}
//
//
// // Package markparsr provides utilities for validating markdown documentation
// // for Terraform modules, specifically focusing on README files and ensuring they
// // match the actual Terraform code.
// package markparsr
//
// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// )
//
// // Validator is an interface for all validators
// type Validator interface {
// 	Validate() []error
// }
//
// // MarkdownValidator orchestrates all validations
// type MarkdownValidator struct {
// 	ReadmePath string
// 	Data       string
// 	validators []Validator
// }
//
// // Config holds configuration options for the validator
// type Config struct {
// 	// ReadmePath is the path to the README.md file
// 	ReadmePath string
//
// 	// SkipURLValidation skips validation of URLs in the markdown
// 	SkipURLValidation bool
//
// 	// SkipFileValidation skips validation of required files
// 	SkipFileValidation bool
//
// 	// SkipTerraformValidation skips validation of Terraform definitions
// 	SkipTerraformValidation bool
//
// 	// SkipVariablesValidation skips validation of Terraform variables
// 	SkipVariablesValidation bool
//
// 	// SkipOutputsValidation skips validation of Terraform outputs
// 	SkipOutputsValidation bool
// }
//
// // DefaultConfig returns a default configuration
// func DefaultConfig() *Config {
// 	return &Config{
// 		ReadmePath: "README.md",
// 	}
// }
//
// // New creates a new MarkdownValidator with the given configuration
// func New(config *Config) (*MarkdownValidator, error) {
// 	if config == nil {
// 		config = DefaultConfig()
// 	}
//
// 	// Allow overriding the README path via environment variable
// 	readmePath := config.ReadmePath
// 	if envPath := os.Getenv("README_PATH"); envPath != "" {
// 		readmePath = envPath
// 	}
//
// 	absReadmePath, err := filepath.Abs(readmePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get absolute path: %v", err)
// 	}
//
// 	dataBytes, err := os.ReadFile(absReadmePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read file: %v", err)
// 	}
// 	data := string(dataBytes)
//
// 	mv := &MarkdownValidator{
// 		ReadmePath: absReadmePath,
// 		Data:       data,
// 	}
//
// 	// Initialize validators based on configuration
// 	mv.validators = []Validator{
// 		NewSectionValidator(data),
// 	}
//
// 	if !config.SkipFileValidation {
// 		mv.validators = append(mv.validators, NewFileValidator(absReadmePath))
// 	}
//
// 	if !config.SkipURLValidation {
// 		mv.validators = append(mv.validators, NewURLValidator(data))
// 	}
//
// 	if !config.SkipTerraformValidation {
// 		mv.validators = append(mv.validators, NewTerraformDefinitionValidator(data))
// 	}
//
// 	if !config.SkipVariablesValidation {
// 		mv.validators = append(mv.validators, NewItemValidator(data, "Variables", "variable", "Inputs", "variables.tf"))
// 	}
//
// 	if !config.SkipOutputsValidation {
// 		mv.validators = append(mv.validators, NewItemValidator(data, "Outputs", "output", "Outputs", "outputs.tf"))
// 	}
//
// 	return mv, nil
// }
//
// // Validate runs all registered validators
// func (mv *MarkdownValidator) Validate() []error {
// 	var allErrors []error
// 	for _, validator := range mv.validators {
// 		allErrors = append(allErrors, validator.Validate()...)
// 	}
// 	return allErrors
// }
//
// // ValidateWithFailFast runs validators and returns on first error
// func (mv *MarkdownValidator) ValidateWithFailFast() error {
// 	for _, validator := range mv.validators {
// 		errors := validator.Validate()
// 		if len(errors) > 0 {
// 			return errors[0]
// 		}
// 	}
// 	return nil
// }
