package markparsr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTerraformDefinitionValidator(t *testing.T) {
	mc := NewMarkdownContent("", FormatDocument, nil)
	tc, _ := NewTerraformContent("")

	tdv := NewTerraformDefinitionValidator(mc, tc)

	if tdv == nil {
		t.Fatal("NewTerraformDefinitionValidator() returned nil")
	}

	if tdv.markdown == nil {
		t.Error("NewTerraformDefinitionValidator() markdown is nil")
	}

	if tdv.terraform == nil {
		t.Error("NewTerraformDefinitionValidator() terraform is nil")
	}
}

func TestTerraformDefinitionValidator_Validate(t *testing.T) {
	// Create a temporary directory for each test to write actual files
	createTestModule := func(terraformData string) (string, error) {
		tmpDir, err := os.MkdirTemp("", "test-module-*")
		if err != nil {
			return "", err
		}

		if terraformData != "" {
			mainPath := filepath.Join(tmpDir, "main.tf")
			if err := os.WriteFile(mainPath, []byte(terraformData), 0o644); err != nil {
				os.RemoveAll(tmpDir)
				return "", err
			}
		}

		return tmpDir, nil
	}

	tests := []struct {
		name           string
		markdownData   string
		terraformData  string
		providerPrefix []string
		expectedErrors int
		errorContains  []string
	}{
		{
			name: "matching resources and data sources",
			markdownData: `## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
- [azurerm_virtual_network](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network)

Data sources:
- [azurerm_client_config](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/client_config)
`,
			terraformData: `
resource "azurerm_resource_group" "main" {
  name     = "test-rg"
  location = "westeurope"
}

resource "azurerm_virtual_network" "main" {
  name = "test-vnet"
}

data "azurerm_client_config" "current" {}
`,
			providerPrefix: []string{"azurerm_"},
			expectedErrors: 0,
		},
		{
			name: "missing resource in markdown",
			markdownData: `## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
`,
			terraformData: `
resource "azurerm_resource_group" "main" {
  name = "test-rg"
}

resource "azurerm_virtual_network" "main" {
  name = "test-vnet"
}
`,
			providerPrefix: []string{"azurerm_"},
			expectedErrors: 1,
			errorContains:  []string{"azurerm_virtual_network", "missing in markdown"},
		},
		{
			name: "extra resource in markdown",
			markdownData: `## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
- [azurerm_virtual_network](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network)
`,
			terraformData: `
resource "azurerm_resource_group" "main" {
  name = "test-rg"
}
`,
			providerPrefix: []string{"azurerm_"},
			expectedErrors: 1,
			errorContains:  []string{"azurerm_virtual_network", "missing in Terraform"},
		},
		{
			name: "missing data source in markdown",
			markdownData: `## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
`,
			terraformData: `
resource "azurerm_resource_group" "main" {
  name = "test-rg"
}

data "azurerm_client_config" "current" {}
`,
			providerPrefix: []string{"azurerm_"},
			expectedErrors: 1,
			errorContains:  []string{"azurerm_client_config", "missing in markdown"},
		},
		{
			name:         "no resources section in markdown but resources in terraform",
			markdownData: `## Inputs\n\nSome inputs`,
			terraformData: `
resource "azurerm_resource_group" "main" {
  name = "test-rg"
}
`,
			providerPrefix: []string{"azurerm_"},
			expectedErrors: 1,
			errorContains:  []string{"resources section not found"},
		},
		{
			name:           "no resources in both",
			markdownData:   `## Something\n\ncontent`,
			terraformData:  "",
			providerPrefix: []string{"azurerm_"},
			expectedErrors: 0,
		},
		{
			name: "qualified resource names",
			markdownData: `## Resources

- [azurerm_resource_group.main](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
`,
			terraformData: `
resource "azurerm_resource_group" "main" {
  name = "test-rg"
}
`,
			providerPrefix: []string{"azurerm_"},
			expectedErrors: 0,
		},
		{
			name: "resources section exists but is empty",
			markdownData: `## Resources

No resources here
`,
			terraformData: `
resource "azurerm_resource_group" "main" {
  name = "test-rg"
}
`,
			providerPrefix: []string{"azurerm_"},
			expectedErrors: 1,
			errorContains:  []string{"resources section not found or empty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMarkdownContent(tt.markdownData, FormatDocument, tt.providerPrefix)

			tmpDir, err := createTestModule(tt.terraformData)
			if err != nil {
				t.Fatalf("Failed to create test module: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			tc, err := NewTerraformContent(tmpDir)
			if err != nil {
				t.Fatalf("Failed to create TerraformContent: %v", err)
			}

			tdv := NewTerraformDefinitionValidator(mc, tc)
			errs := tdv.Validate()

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

func TestTerraformDefinitionValidator_TerraformExtractionError(t *testing.T) {
	mc := NewMarkdownContent("## Resources\n\n- [azurerm_resource_group](https://example.com)", FormatDocument, []string{"azurerm_"})

	tmpDir := t.TempDir()
	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	tdv := NewTerraformDefinitionValidator(mc, tc)
	errs := tdv.Validate()

	if len(errs) != 1 {
		t.Errorf("Validate() with terraform extraction error returned %d errors; want 1", len(errs))
		for _, err := range errs {
			t.Logf("  error: %v", err)
		}
	}
}

func TestTerraformDefinitionValidator_NoResourcesInTerraform(t *testing.T) {
	mc := NewMarkdownContent(`## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
`, FormatDocument, []string{"azurerm_"})

	tmpDir := t.TempDir()
	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	tdv := NewTerraformDefinitionValidator(mc, tc)
	errs := tdv.Validate()

	if len(errs) != 1 {
		t.Errorf("Validate() with no resources in terraform returned %d errors; want 1", len(errs))
		for _, err := range errs {
			t.Logf("  error: %v", err)
		}
	}
}

func TestTerraformDefinitionValidator_MultipleResources(t *testing.T) {
	markdownData := `## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
- [azurerm_virtual_network](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network)
- [azurerm_subnet](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/subnet)

Data sources:
- [azurerm_client_config](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/client_config)
- [azurerm_subscription](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/subscription)
`

	terraformData := `
resource "azurerm_resource_group" "main" {
  name = "test-rg"
}

resource "azurerm_virtual_network" "main" {
  name = "test-vnet"
}

resource "azurerm_subnet" "main" {
  name = "test-subnet"
}

data "azurerm_client_config" "current" {}

data "azurerm_subscription" "current" {}
`

	mc := NewMarkdownContent(markdownData, FormatDocument, []string{"azurerm_"})

	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.tf")
	if err := os.WriteFile(mainPath, []byte(terraformData), 0o644); err != nil {
		t.Fatalf("Failed to write terraform file: %v", err)
	}

	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	tdv := NewTerraformDefinitionValidator(mc, tc)
	errs := tdv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with multiple resources returned %d errors; want 0", len(errs))
		for _, err := range errs {
			t.Logf("  error: %v", err)
		}
	}
}

func TestTerraformDefinitionValidator_OnlyCheckWhenRelevant(t *testing.T) {
	t.Run("no validation when no resources in terraform and no section", func(t *testing.T) {
		mc := NewMarkdownContent("## Inputs\n\nSome inputs", FormatDocument, []string{"azurerm_"})
		tmpDir := t.TempDir()
		tc, err := NewTerraformContent(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create TerraformContent: %v", err)
		}

		tdv := NewTerraformDefinitionValidator(mc, tc)
		errs := tdv.Validate()

		if len(errs) != 0 {
			t.Errorf("Validate() should not validate when no resources exist in terraform and no section in markdown; got %d errors", len(errs))
		}
	})

	t.Run("validates when resources section exists in markdown", func(t *testing.T) {
		mc := NewMarkdownContent("## Resources\n\n- [azurerm_rg](https://example.com)", FormatDocument, []string{"azurerm_"})
		tmpDir := t.TempDir()
		tc, err := NewTerraformContent(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create TerraformContent: %v", err)
		}

		tdv := NewTerraformDefinitionValidator(mc, tc)
		errs := tdv.Validate()

		if len(errs) == 0 {
			t.Error("Validate() should validate when resources section exists in markdown")
		}
	})
}

func TestTerraformDefinitionValidator_MixedQualifiedAndBase(t *testing.T) {
	markdownData := `## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
- [azurerm_virtual_network.main](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network)
`

	terraformData := `
resource "azurerm_resource_group" "rg" {
  name = "test-rg"
}

resource "azurerm_virtual_network" "main" {
  name = "test-vnet"
}
`

	mc := NewMarkdownContent(markdownData, FormatDocument, []string{"azurerm_"})

	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.tf")
	if err := os.WriteFile(mainPath, []byte(terraformData), 0o644); err != nil {
		t.Fatalf("Failed to write terraform file: %v", err)
	}

	tc, err := NewTerraformContent(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create TerraformContent: %v", err)
	}

	tdv := NewTerraformDefinitionValidator(mc, tc)
	errs := tdv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with mixed qualified and base names returned %d errors; want 0", len(errs))
		for _, err := range errs {
			t.Logf("  error: %v", err)
		}
	}
}
