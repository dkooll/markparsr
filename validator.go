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

	// AdditionalSections specifies section names that should exist in the markdown
	// If empty, only the required sections will be validated.
	AdditionalSections []string

	// AdditionalFiles specifies additional files that should exist
	// These can be relative paths (to the module directory) or absolute paths
	AdditionalFiles []string

	// ReadmePath specifies the path to the README file
	// If empty, README_PATH environment variable will be used
	ReadmePath string

	// ProviderPrefixes specifies the provider prefixes to recognize in resources
	ProviderPrefixes []string
}

// Option is a function that configures Options
type Option func(*Options)

// WithFormat sets the markdown format explicitly
func WithFormat(format MarkdownFormat) Option {
	return func(o *Options) {
		o.Format = format
	}
}

// WithAdditionalSections specifies additional sections to validate
func WithAdditionalSections(sections ...string) Option {
	return func(o *Options) {
		o.AdditionalSections = sections
	}
}

// WithAdditionalFiles specifies additional files to validate
func WithAdditionalFiles(files ...string) Option {
	return func(o *Options) {
		o.AdditionalFiles = files
	}
}

// WithRelativeReadmePath specifies the path to the README file
func WithRelativeReadmePath(path string) Option {
	return func(o *Options) {
		o.ReadmePath = path
	}
}

// WithProviderPrefixes specifies custom provider prefixes to recognize
func WithProviderPrefixes(prefixes ...string) Option {
	return func(o *Options) {
		o.ProviderPrefixes = prefixes
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

// NewReadmeValidator creates a validator with the given options.
// If no readme path is provided via WithRelativeReadmePath, the README_PATH environment variable will be used.
func NewReadmeValidator(opts ...Option) (*ReadmeValidator, error) {
	// Initialize with default options
	options := Options{
		Format:             FormatAuto,
		AdditionalSections: []string{},
		AdditionalFiles:    []string{},
		ReadmePath:         "",
		ProviderPrefixes:   []string{},
	}

	// Apply all functional options
	for _, opt := range opts {
		opt(&options)
	}

	// Check for format override in environment variables
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

	// Determine the README path - first check options, then environment
	var finalReadmePath string
	if options.ReadmePath != "" {
		finalReadmePath = options.ReadmePath
	} else {
		finalReadmePath = os.Getenv("README_PATH")
		if finalReadmePath == "" {
			return nil, fmt.Errorf("README path not provided via WithRelativeReadmePath and README_PATH environment variable not set")
		}
		if os.Getenv("VERBOSE") == "true" {
			fmt.Printf("Using README_PATH from environment: %s\n", finalReadmePath)
		}
	}

	// Convert to absolute path
	readmeFile, err := filepath.Abs(finalReadmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for README: %w", err)
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

	// Initialize content analyzers
	markdown := NewMarkdownContent(string(data), options.Format, options.ProviderPrefixes)

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
		NewSectionValidator(markdown, options.AdditionalSections),
		NewFileValidator(readmeFile, absModulePath, options.AdditionalFiles),
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
