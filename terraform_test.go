package markparsr

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type mockFileReader struct {
	files map[string][]byte
	err   error
}

func (m *mockFileReader) ReadFile(path string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	if content, ok := m.files[path]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

type mockHCLParser struct {
	files map[string]*hcl.File
	diags hcl.Diagnostics
}

type mockDirEntry struct {
	name string
	dir  bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.dir }
func (m mockDirEntry) Type() fs.FileMode          { return 0 }
func (m mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func (m *mockHCLParser) ParseHCL(content []byte, filename string) (*hcl.File, hcl.Diagnostics) {
	if m.diags.HasErrors() {
		return nil, m.diags
	}
	if file, ok := m.files[filename]; ok {
		return file, nil
	}
	return &hcl.File{Body: hcl.EmptyBody()}, nil
}

func TestNewTerraformContent(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		wantErr    bool
	}{
		{
			name:       "with valid path",
			modulePath: "/tmp/test-module",
			wantErr:    false,
		},
		{
			name:       "with empty path uses current dir",
			modulePath: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc, err := NewTerraformContent(tt.modulePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTerraformContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tc == nil {
				t.Error("NewTerraformContent() returned nil without error")
			}
		})
	}
}

func TestTerraformContent_ExtractItems(t *testing.T) {
	variableHCL := `
variable "name" {
  type = string
}

variable "location" {
  type = string
}
`
	outputHCL := `
output "id" {
  value = azurerm_resource.main.id
}

output "name" {
  value = azurerm_resource.main.name
}
`

	tests := []struct {
		name          string
		filePath      string
		blockType     string
		fileContent   string
		expectedItems []string
		expectError   bool
	}{
		{
			name:          "extract variables",
			filePath:      "variables.tf",
			blockType:     "variable",
			fileContent:   variableHCL,
			expectedItems: []string{"name", "location"},
			expectError:   false,
		},
		{
			name:          "extract outputs",
			filePath:      "outputs.tf",
			blockType:     "output",
			fileContent:   outputHCL,
			expectedItems: []string{"id", "name"},
			expectError:   false,
		},
		{
			name:          "file not found returns empty",
			filePath:      "nonexistent.tf",
			blockType:     "variable",
			fileContent:   "",
			expectedItems: []string{},
			expectError:   false,
		},
		{
			name:          "empty file returns empty",
			filePath:      "empty.tf",
			blockType:     "variable",
			fileContent:   "",
			expectedItems: []string{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TerraformContent{
				workspace: "/tmp/test",
			}

			if tt.fileContent != "" {
				file, diags := hclsyntax.ParseConfig([]byte(tt.fileContent), tt.filePath, hcl.Pos{Line: 1, Column: 1})
				if diags.HasErrors() {
					t.Fatalf("Failed to parse test HCL: %v", diags)
				}

				tc.fileReader = &mockFileReader{
					files: map[string][]byte{
						tt.filePath: []byte(tt.fileContent),
					},
				}
				tc.hclParser = &mockHCLParser{
					files: map[string]*hcl.File{
						tt.filePath: file,
					},
				}
			} else {
				tc.fileReader = &mockFileReader{
					files: map[string][]byte{},
				}
				tc.hclParser = &mockHCLParser{
					files: map[string]*hcl.File{},
				}
			}

			items, err := tc.ExtractItems(tt.filePath, tt.blockType)
			if (err != nil) != tt.expectError {
				t.Errorf("ExtractItems() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if len(items) != len(tt.expectedItems) {
				t.Errorf("ExtractItems() returned %d items; want %d", len(items), len(tt.expectedItems))
				t.Logf("Got: %v", items)
				t.Logf("Want: %v", tt.expectedItems)
				return
			}

			for _, expected := range tt.expectedItems {
				if !slices.Contains(items, expected) {
					t.Errorf("ExtractItems() missing expected item %q", expected)
				}
			}
		})
	}
}

func TestTerraformContent_ExtractResourcesAndDataSources(t *testing.T) {
	mainHCL := `
resource "azurerm_resource_group" "main" {
  name     = "test-rg"
  location = "westeurope"
}

resource "azurerm_virtual_network" "main" {
  name                = "test-vnet"
  resource_group_name = azurerm_resource_group.main.name
}

data "azurerm_client_config" "current" {}
`

	tests := []struct {
		name              string
		fileContent       string
		expectedResources []string
		expectedDataSrcs  []string
		expectError       bool
	}{
		{
			name:        "extract resources and data sources",
			fileContent: mainHCL,
			expectedResources: []string{
				"azurerm_resource_group",
				"azurerm_resource_group.main",
				"azurerm_virtual_network",
				"azurerm_virtual_network.main",
			},
			expectedDataSrcs: []string{
				"azurerm_client_config",
				"azurerm_client_config.current",
			},
			expectError: false,
		},
		{
			name:              "empty file returns empty lists",
			fileContent:       "",
			expectedResources: []string{},
			expectedDataSrcs:  []string{},
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TerraformContent{
				workspace: "/tmp/test",
			}

			if tt.fileContent != "" {
				file, diags := hclsyntax.ParseConfig([]byte(tt.fileContent), "main.tf", hcl.Pos{Line: 1, Column: 1})
				if diags.HasErrors() {
					t.Fatalf("Failed to parse test HCL: %v", diags)
				}

				tc.fileReader = &mockFileReader{
					files: map[string][]byte{
						filepath.Join("/tmp/test", "main.tf"): []byte(tt.fileContent),
					},
				}
				tc.hclParser = &mockHCLParser{
					files: map[string]*hcl.File{
						filepath.Join("/tmp/test", "main.tf"): file,
					},
				}
				tc.readDir = func(path string) ([]os.DirEntry, error) {
					return []os.DirEntry{mockDirEntry{name: "main.tf"}}, nil
				}
			} else {
				tc.fileReader = &mockFileReader{files: map[string][]byte{}}
				tc.hclParser = &mockHCLParser{files: map[string]*hcl.File{}}
				tc.readDir = func(path string) ([]os.DirEntry, error) {
					return []os.DirEntry{}, nil
				}
			}

			resources, dataSources, err := tc.ExtractResourcesAndDataSources()
			if (err != nil) != tt.expectError {
				t.Errorf("ExtractResourcesAndDataSources() error = %v, expectError %v", err, tt.expectError)
				return
			}

			for _, expected := range tt.expectedResources {
				if !slices.Contains(resources, expected) {
					t.Errorf("ExtractResourcesAndDataSources() missing expected resource %q", expected)
				}
			}

			for _, expected := range tt.expectedDataSrcs {
				if !slices.Contains(dataSources, expected) {
					t.Errorf("ExtractResourcesAndDataSources() missing expected data source %q", expected)
				}
			}
		})
	}
}

func TestTerraformContent_ParseFile(t *testing.T) {
	validHCL := `variable "test" { type = string }`
	invalidHCL := `variable "test" { invalid syntax`

	tests := []struct {
		name        string
		filePath    string
		fileContent string
		setupMock   func() (*mockFileReader, *mockHCLParser)
		expectError bool
	}{
		{
			name:        "parse valid file",
			filePath:    "test.tf",
			fileContent: validHCL,
			setupMock: func() (*mockFileReader, *mockHCLParser) {
				file, _ := hclsyntax.ParseConfig([]byte(validHCL), "test.tf", hcl.Pos{Line: 1, Column: 1})
				return &mockFileReader{
						files: map[string][]byte{"test.tf": []byte(validHCL)},
					}, &mockHCLParser{
						files: map[string]*hcl.File{"test.tf": file},
					}
			},
			expectError: false,
		},
		{
			name:     "file not found returns nil",
			filePath: "notfound.tf",
			setupMock: func() (*mockFileReader, *mockHCLParser) {
				return &mockFileReader{
						files: map[string][]byte{},
					}, &mockHCLParser{
						files: map[string]*hcl.File{},
					}
			},
			expectError: false,
		},
		{
			name:        "parse error returns error",
			filePath:    "invalid.tf",
			fileContent: invalidHCL,
			setupMock: func() (*mockFileReader, *mockHCLParser) {
				return &mockFileReader{
						files: map[string][]byte{"invalid.tf": []byte(invalidHCL)},
					}, &mockHCLParser{
						diags: hcl.Diagnostics{
							{
								Severity: hcl.DiagError,
								Summary:  "Parse error",
							},
						},
					}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr, hp := tt.setupMock()
			tc := &TerraformContent{
				workspace:  "/tmp/test",
				fileReader: fr,
				hclParser:  hp,
			}

			file, err := tc.parseFile(tt.filePath)
			if (err != nil) != tt.expectError {
				t.Errorf("parseFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && err == nil && file == nil && tt.fileContent != "" {
				t.Error("parseFile() returned nil file without error for existing file")
			}
		})
	}
}

func TestDefaultFileReader_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tf")
	testContent := []byte("variable \"test\" { type = string }")

	if err := os.WriteFile(testFile, testContent, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fr := &defaultFileReader{}

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "read existing file",
			path:        testFile,
			expectError: false,
		},
		{
			name:        "read non-existing file",
			path:        filepath.Join(tmpDir, "nonexistent.tf"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := fr.ReadFile(tt.path)
			if (err != nil) != tt.expectError {
				t.Errorf("ReadFile() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && len(content) == 0 {
				t.Error("ReadFile() returned empty content for existing file")
			}
		})
	}
}

func TestDefaultHCLParser_ParseHCL(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "parse valid HCL",
			content:     `variable "test" { type = string }`,
			expectError: false,
		},
		{
			name:        "parse invalid HCL",
			content:     `variable "test" { invalid`,
			expectError: true,
		},
		{
			name:        "parse empty content",
			content:     "",
			expectError: false,
		},
	}

	hp := &defaultHCLParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, diags := hp.ParseHCL([]byte(tt.content), "test.tf")
			if diags.HasErrors() != tt.expectError {
				t.Errorf("ParseHCL() hasErrors = %v, expectError %v", diags.HasErrors(), tt.expectError)
			}
			if !tt.expectError && file == nil {
				t.Error("ParseHCL() returned nil file without errors")
			}
		})
	}
}

func ExampleTerraformContent_ExtractItems() {
	tmpDir := os.TempDir()
	varsFile := filepath.Join(tmpDir, "example_vars.tf")

	content := `
variable "name" {
  type = string
}

variable "location" {
  type = string
}
`
	if err := os.WriteFile(varsFile, []byte(content), 0o644); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer os.Remove(varsFile)

	tc, _ := NewTerraformContent(tmpDir)
	items, _ := tc.ExtractItems(varsFile, "variable")

	for _, item := range items {
		fmt.Println(item)
	}
}
