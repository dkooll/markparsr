package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Format             MarkdownFormat
	AdditionalSections []string
	AdditionalFiles    []string
	ReadmePath         string
	ProviderPrefixes   []string
}

type Option func(*Options)

func WithFormat(format MarkdownFormat) Option {
	return func(o *Options) {
		o.Format = format
	}
}

func WithAdditionalSections(sections ...string) Option {
	return func(o *Options) {
		o.AdditionalSections = sections
	}
}

func WithAdditionalFiles(files ...string) Option {
	return func(o *Options) {
		o.AdditionalFiles = files
	}
}

func WithRelativeReadmePath(path string) Option {
	return func(o *Options) {
		o.ReadmePath = path
	}
}

func WithProviderPrefixes(prefixes ...string) Option {
	return func(o *Options) {
		o.ProviderPrefixes = prefixes
	}
}

type ReadmeValidator struct {
	readmePath string
	modulePath string
	markdown   *MarkdownContent
	terraform  *TerraformContent
	validators []Validator
	options    Options
}

func NewReadmeValidator(opts ...Option) (*ReadmeValidator, error) {
	options := Options{
		Format:             FormatDocument,
		AdditionalSections: []string{},
		AdditionalFiles:    []string{},
		ReadmePath:         "",
		ProviderPrefixes:   []string{},
	}

	for _, opt := range opts {
		opt(&options)
	}

	if envFormat := os.Getenv("FORMAT"); envFormat != "" {
		switch strings.ToLower(envFormat) {
		case "document":
			options.Format = FormatDocument
		case "table":
			fmt.Println("Table format is no longer supported; defaulting to document format")
			options.Format = FormatDocument
		default:
			fmt.Printf("Unknown format in FORMAT environment variable: %s, using document format\n", envFormat)
		}
	}

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

	readmeFile, err := filepath.Abs(finalReadmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for README: %w", err)
	}

	modulePath := filepath.Dir(readmeFile)
	if envModulePath := os.Getenv("MODULE_PATH"); envModulePath != "" {
		modulePath = envModulePath
	}

	absModulePath, err := filepath.Abs(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute module path: %w", err)
	}

	data, err := os.ReadFile(readmeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	markdown := NewMarkdownContent(string(data), options.Format, options.ProviderPrefixes)

	terraform, err := NewTerraformContent(absModulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terraform content: %w", err)
	}

	validator := &ReadmeValidator{
		readmePath: readmeFile,
		modulePath: absModulePath,
		markdown:   markdown,
		terraform:  terraform,
		options:    options,
	}

	validator.validators = buildDefaultValidators(readmeFile, absModulePath, markdown, terraform, options)

	return validator, nil
}

func buildDefaultValidators(readmePath, modulePath string, markdown *MarkdownContent, terraform *TerraformContent, options Options) []Validator {
	return []Validator{
		NewSectionValidator(markdown, options.AdditionalSections),
		NewFileValidator(readmePath, modulePath, options.AdditionalFiles),
		NewURLValidator(markdown),
		NewTerraformDefinitionValidator(markdown, terraform),
		NewItemValidator(markdown, terraform, "Variables", "variable", []string{"Required Inputs", "Optional Inputs"}, "variables.tf"),
		NewItemValidator(markdown, terraform, "Outputs", "output", []string{"Outputs"}, "outputs.tf"),
	}
}

func (rv *ReadmeValidator) Validate() []error {
	collector := &ErrorCollector{}

	for _, validator := range rv.validators {
		collector.AddMany(validator.Validate())
	}

	return collector.Errors()
}

func (rv *ReadmeValidator) GetFormat() MarkdownFormat {
	if rv.markdown != nil {
		return rv.markdown.format
	}
	return FormatDocument
}
