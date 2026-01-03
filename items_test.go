package markparsr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewItemValidator(t *testing.T) {
	markdown := NewMarkdownContent("", FormatDocument, nil)
	terraform, _ := NewTerraformContent("")

	tests := []struct {
		name      string
		itemType  string
		blockType string
		sections  []string
		fileName  string
	}{
		{
			name:      "variables validator",
			itemType:  "Variables",
			blockType: "variable",
			sections:  []string{"Required Inputs", "Optional Inputs"},
			fileName:  "variables.tf",
		},
		{
			name:      "outputs validator",
			itemType:  "Outputs",
			blockType: "output",
			sections:  []string{"Outputs"},
			fileName:  "outputs.tf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iv := NewItemValidator(markdown, terraform, tt.itemType, tt.blockType, tt.sections, tt.fileName)

			if iv == nil {
				t.Fatal("NewItemValidator() returned nil")
			}

			if iv.itemType != tt.itemType {
				t.Errorf("NewItemValidator() itemType = %q; want %q", iv.itemType, tt.itemType)
			}

			if iv.blockType != tt.blockType {
				t.Errorf("NewItemValidator() blockType = %q; want %q", iv.blockType, tt.blockType)
			}

			if len(iv.sections) != len(tt.sections) {
				t.Errorf("NewItemValidator() sections count = %d; want %d", len(iv.sections), len(tt.sections))
			}

			if iv.fileName != tt.fileName {
				t.Errorf("NewItemValidator() fileName = %q; want %q", iv.fileName, tt.fileName)
			}
		})
	}
}

func TestItemValidator_Validate(t *testing.T) {
	tests := []struct {
		name           string
		markdownData   string
		terraformData  string
		itemType       string
		blockType      string
		sections       []string
		expectedErrors int
		errorContains  []string
	}{
		{
			name: "matching variables",
			markdownData: `## Required Inputs

### <a name="input_name"></a> name

Description

### <a name="input_location"></a> location

Description
`,
			terraformData: `
variable "name" {
  type = string
}

variable "location" {
  type = string
}
`,
			itemType:       "Variables",
			blockType:      "variable",
			sections:       []string{"Required Inputs"},
			expectedErrors: 0,
		},
		{
			name: "missing variable in markdown",
			markdownData: `## Required Inputs

### <a name="input_name"></a> name

Description
`,
			terraformData: `
variable "name" {
  type = string
}

variable "location" {
  type = string
}
`,
			itemType:       "Variables",
			blockType:      "variable",
			sections:       []string{"Required Inputs"},
			expectedErrors: 1,
			errorContains:  []string{"location", "missing in markdown"},
		},
		{
			name: "extra variable in markdown",
			markdownData: `## Required Inputs

### <a name="input_name"></a> name

Description

### <a name="input_location"></a> location

Description
`,
			terraformData: `
variable "name" {
  type = string
}
`,
			itemType:       "Variables",
			blockType:      "variable",
			sections:       []string{"Required Inputs"},
			expectedErrors: 1,
			errorContains:  []string{"location", "missing in Terraform"},
		},
		{
			name:           "no section and no terraform items",
			markdownData:   `## Something Else\n\ncontent`,
			terraformData:  "",
			itemType:       "Outputs",
			blockType:      "output",
			sections:       []string{"Outputs"},
			expectedErrors: 0,
		},
		{
			name: "section present but empty",
			markdownData: `## Outputs

No items here
`,
			terraformData: `
output "id" {
  value = "test"
}
`,
			itemType:       "Outputs",
			blockType:      "output",
			sections:       []string{"Outputs"},
			expectedErrors: 1,
			errorContains:  []string{"id", "missing in markdown"},
		},
		{
			name: "multiple sections combined",
			markdownData: `## Required Inputs

### <a name="input_name"></a> name

Description

## Optional Inputs

### <a name="input_location"></a> location

Description
`,
			terraformData: `
variable "name" {
  type = string
}

variable "location" {
  type = string
  default = "westeurope"
}
`,
			itemType:       "Variables",
			blockType:      "variable",
			sections:       []string{"Required Inputs", "Optional Inputs"},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMarkdownContent(tt.markdownData, FormatDocument, nil)

			tmpDir := t.TempDir()
			if tt.terraformData != "" {
				testPath := filepath.Join(tmpDir, "test.tf")
				if err := os.WriteFile(testPath, []byte(tt.terraformData), 0o644); err != nil {
					t.Fatalf("Failed to write terraform file: %v", err)
				}
			}

			tc, err := NewTerraformContent(tmpDir)
			if err != nil {
				t.Fatalf("Failed to create TerraformContent: %v", err)
			}

			iv := NewItemValidator(mc, tc, tt.itemType, tt.blockType, tt.sections, "test.tf")
			errs := iv.Validate()

			if len(errs) != tt.expectedErrors {
				t.Errorf("Validate() returned %d errors; want %d", len(errs), tt.expectedErrors)
				for i, err := range errs {
					t.Logf("  error %d: %v", i+1, err)
				}
			}

			for _, substr := range tt.errorContains {
				found := false
				for _, err := range errs {
					if err != nil && strings.Contains(err.Error(), substr) {
						found = true
						break
					}
				}
				if !found && tt.expectedErrors > 0 {
					t.Errorf("Expected error containing %q, but not found", substr)
				}
			}
		})
	}
}

func TestItemValidator_WithFallbackAnchors(t *testing.T) {
	markdownData := `Some content

<a name="input_var1"></a>
<a name="input_var2"></a>
`
	terraformData := `
variable "var1" {
  type = string
}

variable "var2" {
  type = string
}
`

	mc := NewMarkdownContent(markdownData, FormatDocument, nil)

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.tf")
	if err := os.WriteFile(testPath, []byte(terraformData), 0o644); err != nil {
		t.Fatalf("Failed to write terraform file: %v", err)
	}

	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	iv := NewItemValidator(mc, tc, "Variables", "variable", []string{"Required Inputs"}, "test.tf")
	errs := iv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with anchor fallback returned %d errors; want 0", len(errs))
		for _, err := range errs {
			t.Logf("  error: %v", err)
		}
	}
}

func TestItemValidator_EmptyBoth(t *testing.T) {
	mc := NewMarkdownContent("", FormatDocument, nil)

	tmpDir := t.TempDir()
	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	iv := NewItemValidator(mc, tc, "Variables", "variable", []string{"Required Inputs"}, "test.tf")
	errs := iv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with no items on both sides returned %d errors; want 0", len(errs))
	}
}

func TestItemValidator_TerraformExtractionError(t *testing.T) {
	mc := NewMarkdownContent("## Outputs\n\nSome content", FormatDocument, nil)

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.tf")
	// Write invalid HCL to trigger error
	if err := os.WriteFile(testPath, []byte("invalid hcl {{{"), 0o644); err != nil {
		t.Fatalf("Failed to write invalid terraform file: %v", err)
	}

	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	iv := NewItemValidator(mc, tc, "Outputs", "output", []string{"Outputs"}, testPath)
	errs := iv.Validate()

	if len(errs) != 1 {
		t.Errorf("Validate() with terraform extraction error returned %d errors; want 1", len(errs))
	}
}

func TestItemValidator_MultipleSections(t *testing.T) {
	markdownData := `## Required Inputs

### <a name="input_required1"></a> required1

Required variable

## Optional Inputs

### <a name="input_optional1"></a> optional1

Optional variable

### <a name="input_optional2"></a> optional2

Another optional variable
`
	terraformData := `
variable "required1" {
  type = string
}

variable "optional1" {
  type = string
  default = "default1"
}

variable "optional2" {
  type = string
  default = "default2"
}
`

	mc := NewMarkdownContent(markdownData, FormatDocument, nil)

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.tf")
	if err := os.WriteFile(testPath, []byte(terraformData), 0o644); err != nil {
		t.Fatalf("Failed to write terraform file: %v", err)
	}

	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	iv := NewItemValidator(mc, tc, "Variables", "variable", []string{"Required Inputs", "Optional Inputs"}, "test.tf")
	errs := iv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with multiple sections returned %d errors; want 0", len(errs))
		for _, err := range errs {
			t.Logf("  error: %v", err)
		}
	}
}

func TestItemValidator_CaseInsensitive(t *testing.T) {
	markdownData := `## Outputs

### <a name="output_ID"></a> ID

The resource ID
`
	terraformData := `
output "id" {
  value = "test-id"
}
`

	mc := NewMarkdownContent(markdownData, FormatDocument, nil)

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.tf")
	if err := os.WriteFile(testPath, []byte(terraformData), 0o644); err != nil {
		t.Fatalf("Failed to write terraform file: %v", err)
	}

	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	iv := NewItemValidator(mc, tc, "Outputs", "output", []string{"Outputs"}, "test.tf")
	errs := iv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with case differences returned %d errors; want 0", len(errs))
		for _, err := range errs {
			t.Logf("  error: %v", err)
		}
	}
}
