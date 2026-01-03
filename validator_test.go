package markparsr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewReadmeValidator(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	os.WriteFile(readmePath, []byte("# Test"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte("variable \"test\" {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte("output \"test\" {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte("terraform {}"), 0o644)

	tests := []struct {
		name        string
		opts        []Option
		expectError bool
	}{
		{
			name:        "with relative readme path",
			opts:        []Option{WithRelativeReadmePath(readmePath)},
			expectError: false,
		},
		{
			name:        "with format option",
			opts:        []Option{WithRelativeReadmePath(readmePath), WithFormat(FormatDocument)},
			expectError: false,
		},
		{
			name:        "with additional sections",
			opts:        []Option{WithRelativeReadmePath(readmePath), WithAdditionalSections("Examples", "Notes")},
			expectError: false,
		},
		{
			name:        "with additional files",
			opts:        []Option{WithRelativeReadmePath(readmePath), WithAdditionalFiles("main.tf")},
			expectError: false,
		},
		{
			name:        "with provider prefixes",
			opts:        []Option{WithRelativeReadmePath(readmePath), WithProviderPrefixes("azurerm_", "aws_")},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv, err := NewReadmeValidator(tt.opts...)

			if (err != nil) != tt.expectError {
				t.Errorf("NewReadmeValidator() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				if rv == nil {
					t.Fatal("NewReadmeValidator() returned nil without error")
				}

				if rv.markdown == nil {
					t.Error("NewReadmeValidator() markdown is nil")
				}

				if rv.terraform == nil {
					t.Error("NewReadmeValidator() terraform is nil")
				}

				if len(rv.validators) == 0 {
					t.Error("NewReadmeValidator() validators is empty")
				}
			}
		})
	}
}

func TestNewReadmeValidator_Errors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() []Option
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing readme path",
			setup: func() []Option {
				return []Option{}
			},
			expectError: true,
			errorMsg:    "README path not provided",
		},
		{
			name: "non-existent readme file",
			setup: func() []Option {
				return []Option{WithRelativeReadmePath("/nonexistent/README.md")}
			},
			expectError: true,
			errorMsg:    "failed to read file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("README_PATH")

			opts := tt.setup()
			_, err := NewReadmeValidator(opts...)

			if (err != nil) != tt.expectError {
				t.Errorf("NewReadmeValidator() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("NewReadmeValidator() error = %v, should contain %q", err, tt.errorMsg)
			}
		})
	}
}

func TestReadmeValidator_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	readmeContent := `# Test Module

## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)

## Providers

Provider info

## Requirements

Requirement info

## Required Inputs

### <a name="input_name"></a> name

The name

## Optional Inputs

### <a name="input_location"></a> location

The location

## Outputs

### <a name="output_id"></a> id

The ID
`

	os.WriteFile(readmePath, []byte(readmeContent), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(`
variable "name" {
  type = string
}

variable "location" {
  type = string
  default = "westeurope"
}
`), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte(`
output "id" {
  value = "test-id"
}
`), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte(`
terraform {
  required_version = ">= 1.0"
}
`), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(`
resource "azurerm_resource_group" "main" {
  name     = "test-rg"
  location = "westeurope"
}
`), 0o644)

	rv, err := NewReadmeValidator(
		WithRelativeReadmePath(readmePath),
		WithProviderPrefixes("azurerm_"),
	)
	if err != nil {
		t.Fatalf("NewReadmeValidator() failed: %v", err)
	}

	errs := rv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() on valid module returned %d errors; want 0", len(errs))
		for i, err := range errs {
			t.Logf("  error %d: %v", i+1, err)
		}
	}
}

func TestReadmeValidator_ValidateWithErrors(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	readmeContent := `# Test Module

## Resourses

Misspelled section

## Providers

content
`

	os.WriteFile(readmePath, []byte(readmeContent), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(`
variable "name" {
  type = string
}
`), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte{}, 0o644) // Empty file
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte("terraform {}"), 0o644)

	rv, err := NewReadmeValidator(WithRelativeReadmePath(readmePath))
	if err != nil {
		t.Fatalf("NewReadmeValidator() failed: %v", err)
	}

	errs := rv.Validate()

	if len(errs) == 0 {
		t.Error("Validate() on invalid module returned 0 errors; want > 0")
	}

	expectedErrorTypes := []string{
		"misspelled",
		"missing",
		"empty",
	}

	for _, expectedType := range expectedErrorTypes {
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), expectedType) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error type %q not found in errors", expectedType)
		}
	}
}

func TestReadmeValidator_GetFormat(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	os.WriteFile(readmePath, []byte("# Test"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte("variable \"test\" {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte("output \"test\" {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte("terraform {}"), 0o644)

	tests := []struct {
		name           string
		format         MarkdownFormat
		expectedFormat MarkdownFormat
	}{
		{
			name:           "document format",
			format:         FormatDocument,
			expectedFormat: FormatDocument,
		},
		{
			name:           "default format",
			format:         "",
			expectedFormat: FormatDocument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []Option{WithRelativeReadmePath(readmePath)}
			if tt.format != "" {
				opts = append(opts, WithFormat(tt.format))
			}

			rv, err := NewReadmeValidator(opts...)
			if err != nil {
				t.Fatalf("NewReadmeValidator() failed: %v", err)
			}

			format := rv.GetFormat()
			if format != tt.expectedFormat {
				t.Errorf("GetFormat() = %v; want %v", format, tt.expectedFormat)
			}
		})
	}
}

func TestReadmeValidator_EnvironmentVariables(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	os.WriteFile(readmePath, []byte("# Test"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte("variable \"test\" {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte("output \"test\" {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte("terraform {}"), 0o644)

	t.Run("README_PATH env variable", func(t *testing.T) {
		os.Setenv("README_PATH", readmePath)
		defer os.Unsetenv("README_PATH")

		rv, err := NewReadmeValidator()
		if err != nil {
			t.Errorf("NewReadmeValidator() with README_PATH env failed: %v", err)
		}

		if rv == nil {
			t.Fatal("NewReadmeValidator() returned nil")
		}
	})

	t.Run("FORMAT env variable", func(t *testing.T) {
		os.Setenv("README_PATH", readmePath)
		os.Setenv("FORMAT", "document")
		defer func() {
			os.Unsetenv("README_PATH")
			os.Unsetenv("FORMAT")
		}()

		rv, err := NewReadmeValidator()
		if err != nil {
			t.Errorf("NewReadmeValidator() with FORMAT env failed: %v", err)
		}

		if rv == nil {
			t.Fatal("NewReadmeValidator() returned nil")
		}

		if rv.GetFormat() != FormatDocument {
			t.Errorf("GetFormat() = %v; want %v", rv.GetFormat(), FormatDocument)
		}
	})

	t.Run("MODULE_PATH env variable", func(t *testing.T) {
		os.Setenv("README_PATH", readmePath)
		os.Setenv("MODULE_PATH", tmpDir)
		defer func() {
			os.Unsetenv("README_PATH")
			os.Unsetenv("MODULE_PATH")
		}()

		rv, err := NewReadmeValidator()
		if err != nil {
			t.Errorf("NewReadmeValidator() with MODULE_PATH env failed: %v", err)
		}

		if rv == nil {
			t.Fatal("NewReadmeValidator() returned nil")
		}

		if rv.modulePath != tmpDir {
			t.Errorf("modulePath = %q; want %q", rv.modulePath, tmpDir)
		}
	})
}

func TestBuildDefaultValidators(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	mc := NewMarkdownContent("# Test", FormatDocument, nil)
	tc, _ := NewTerraformContent(tmpDir)

	tests := []struct {
		name               string
		additionalSections []string
		additionalFiles    []string
		expectedCount      int
	}{
		{
			name:               "default validators",
			additionalSections: []string{},
			additionalFiles:    []string{},
			expectedCount:      6, // Section, File, URL, TerraformDef, Items(Variables), Items(Outputs)
		},
		{
			name:               "with additional sections",
			additionalSections: []string{"Examples"},
			additionalFiles:    []string{},
			expectedCount:      6,
		},
		{
			name:               "with additional files",
			additionalSections: []string{},
			additionalFiles:    []string{"main.tf"},
			expectedCount:      6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				AdditionalSections: tt.additionalSections,
				AdditionalFiles:    tt.additionalFiles,
			}

			validators := buildDefaultValidators(readmePath, tmpDir, mc, tc, opts)

			if len(validators) != tt.expectedCount {
				t.Errorf("buildDefaultValidators() returned %d validators; want %d", len(validators), tt.expectedCount)
			}
		})
	}
}

func TestReadmeValidator_CompleteIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	readmeContent := `# Complete Test Module

## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
- [azurerm_virtual_network](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network)

## Providers

| Name | Version |
|------|---------|
| azurerm | >= 3.0 |

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |

## Required Inputs

### <a name="input_name"></a> name

The resource name

### <a name="input_location"></a> location

The Azure location

## Optional Inputs

### <a name="input_tags"></a> tags

Resource tags

## Outputs

### <a name="output_id"></a> id

The resource group ID

### <a name="output_vnet_id"></a> vnet_id

The virtual network ID
`

	os.WriteFile(readmePath, []byte(readmeContent), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(`
variable "name" {
  type = string
}

variable "location" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
`), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte(`
output "id" {
  value = azurerm_resource_group.main.id
}

output "vnet_id" {
  value = azurerm_virtual_network.main.id
}
`), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte(`
terraform {
  required_version = ">= 1.0"
}
`), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(`
resource "azurerm_resource_group" "main" {
  name     = var.name
  location = var.location
  tags     = var.tags
}

resource "azurerm_virtual_network" "main" {
  name                = "${var.name}-vnet"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}
`), 0o644)

	rv, err := NewReadmeValidator(
		WithRelativeReadmePath(readmePath),
		WithProviderPrefixes("azurerm_"),
	)
	if err != nil {
		t.Fatalf("NewReadmeValidator() failed: %v", err)
	}

	errs := rv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() on complete valid module returned %d errors; want 0", len(errs))
		for i, err := range errs {
			t.Logf("  error %d: %v", i+1, err)
		}
	}
}

func TestReadmeValidator_Options(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	os.WriteFile(readmePath, []byte("# Test"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte("variable \"test\" {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "outputs.tf"), []byte("output \"test\" {}"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "terraform.tf"), []byte("terraform {}"), 0o644)

	t.Run("WithAdditionalSections", func(t *testing.T) {
		rv, err := NewReadmeValidator(
			WithRelativeReadmePath(readmePath),
			WithAdditionalSections("Examples", "Notes"),
		)
		if err != nil {
			t.Fatalf("NewReadmeValidator() failed: %v", err)
		}

		if len(rv.options.AdditionalSections) != 2 {
			t.Errorf("AdditionalSections count = %d; want 2", len(rv.options.AdditionalSections))
		}
	})

	t.Run("WithAdditionalFiles", func(t *testing.T) {
		rv, err := NewReadmeValidator(
			WithRelativeReadmePath(readmePath),
			WithAdditionalFiles("main.tf", "versions.tf"),
		)
		if err != nil {
			t.Fatalf("NewReadmeValidator() failed: %v", err)
		}

		if len(rv.options.AdditionalFiles) != 2 {
			t.Errorf("AdditionalFiles count = %d; want 2", len(rv.options.AdditionalFiles))
		}
	})

	t.Run("WithProviderPrefixes", func(t *testing.T) {
		rv, err := NewReadmeValidator(
			WithRelativeReadmePath(readmePath),
			WithProviderPrefixes("azurerm_", "aws_", "google_"),
		)
		if err != nil {
			t.Fatalf("NewReadmeValidator() failed: %v", err)
		}

		if len(rv.options.ProviderPrefixes) != 3 {
			t.Errorf("ProviderPrefixes count = %d; want 3", len(rv.options.ProviderPrefixes))
		}
	})
}
