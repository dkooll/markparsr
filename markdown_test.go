package markparsr

import (
	"slices"
	"strings"
	"testing"
)

func TestNewMarkdownContent(t *testing.T) {
	tests := []struct {
		name             string
		data             string
		format           MarkdownFormat
		providerPrefixes []string
		checkSections    []string
		expectSections   []bool
	}{
		{
			name: "basic markdown with sections",
			data: `# Module Title

## Resources

Some resources here

## Providers

Provider info
`,
			format:         FormatDocument,
			checkSections:  []string{"Resources", "Providers"},
			expectSections: []bool{true, true},
		},
		{
			name:           "empty markdown",
			data:           "",
			format:         FormatDocument,
			checkSections:  []string{"Resources"},
			expectSections: []bool{false},
		},
		{
			name: "markdown with h3 items",
			data: `## Required Inputs

### <a name="input_name"></a> name

Description of name
`,
			format:         FormatDocument,
			checkSections:  []string{"Required Inputs"},
			expectSections: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMarkdownContent(tt.data, tt.format, tt.providerPrefixes)

			if mc == nil {
				t.Fatal("NewMarkdownContent() returned nil")
			}

			if mc.GetContent() != tt.data {
				t.Errorf("GetContent() = %q; want %q", mc.GetContent(), tt.data)
			}

			for i, section := range tt.checkSections {
				hasSection := mc.HasSection(section)
				if hasSection != tt.expectSections[i] {
					t.Errorf("HasSection(%q) = %v; want %v", section, hasSection, tt.expectSections[i])
				}
			}
		})
	}
}

func TestMarkdownContent_GetContent(t *testing.T) {
	content := "# Test Content\n\nSome markdown text"
	mc := NewMarkdownContent(content, FormatDocument, nil)

	if mc.GetContent() != content {
		t.Errorf("GetContent() = %q; want %q", mc.GetContent(), content)
	}
}

func TestMarkdownContent_GetAllSections(t *testing.T) {
	data := `# Title

## Resources

content

## Providers

content

## Requirements

content
`
	mc := NewMarkdownContent(data, FormatDocument, nil)
	sections := mc.GetAllSections()

	expectedSections := []string{"Resources", "Providers", "Requirements"}
	if len(sections) != len(expectedSections) {
		t.Errorf("GetAllSections() returned %d sections; want %d", len(sections), len(expectedSections))
	}

	for _, expected := range expectedSections {
		if !slices.Contains(sections, expected) {
			t.Errorf("GetAllSections() missing expected section %q", expected)
		}
	}
}

func TestMarkdownContent_HasSection(t *testing.T) {
	data := `## Resources

Some content

## Providers

More content
`
	mc := NewMarkdownContent(data, FormatDocument, nil)

	tests := []struct {
		name    string
		section string
		want    bool
	}{
		{
			name:    "existing section",
			section: "Resources",
			want:    true,
		},
		{
			name:    "another existing section",
			section: "Providers",
			want:    true,
		},
		{
			name:    "non-existing section",
			section: "Outputs",
			want:    false,
		},
		{
			name:    "case sensitive match",
			section: "resources",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mc.HasSection(tt.section)
			if result != tt.want {
				t.Errorf("HasSection(%q) = %v; want %v", tt.section, result, tt.want)
			}
		})
	}
}

func TestMarkdownContent_ExtractSectionItems(t *testing.T) {
	data := `## Required Inputs

### <a name="input_var1"></a> var1

Description

### <a name="input_var2"></a> var2

Description

## Optional Inputs

### <a name="input_var3"></a> var3

Description
`
	mc := NewMarkdownContent(data, FormatDocument, nil)

	tests := []struct {
		name          string
		sections      []string
		expectedItems []string
	}{
		{
			name:          "extract from single section",
			sections:      []string{"Required Inputs"},
			expectedItems: []string{"var1", "var2"},
		},
		{
			name:          "extract from multiple sections",
			sections:      []string{"Required Inputs", "Optional Inputs"},
			expectedItems: []string{"var1", "var2", "var3"},
		},
		{
			name:          "non-existing section returns empty",
			sections:      []string{"Non Existing"},
			expectedItems: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := mc.ExtractSectionItems(tt.sections...)

			if len(items) != len(tt.expectedItems) {
				t.Errorf("ExtractSectionItems() returned %d items; want %d", len(items), len(tt.expectedItems))
				t.Logf("Got items: %v", items)
				t.Logf("Want items: %v", tt.expectedItems)
				return
			}

			for _, expected := range tt.expectedItems {
				if !slices.Contains(items, expected) {
					t.Errorf("ExtractSectionItems() missing expected item %q", expected)
				}
			}
		})
	}
}

func TestMarkdownContent_ExtractResourcesAndDataSources(t *testing.T) {
	tests := []struct {
		name              string
		data              string
		providerPrefixes  []string
		expectedResources []string
		expectedDataSrcs  []string
		expectError       bool
	}{
		{
			name: "extract resources from resources section",
			data: `## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
- [azurerm_virtual_network](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network)
`,
			providerPrefixes:  []string{"azurerm_"},
			expectedResources: []string{"azurerm_resource_group", "azurerm_virtual_network"},
			expectedDataSrcs:  []string{},
			expectError:       false,
		},
		{
			name: "extract data sources",
			data: `## Resources

- [azurerm_resource_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
- [azurerm_client_config](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/client_config)
`,
			providerPrefixes:  []string{"azurerm_"},
			expectedResources: []string{"azurerm_resource_group"},
			expectedDataSrcs:  []string{"azurerm_client_config"},
			expectError:       false,
		},
		{
			name: "qualified resource names",
			data: `## Resources

- [azurerm_resource_group.main](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
`,
			providerPrefixes:  []string{"azurerm_"},
			expectedResources: []string{"azurerm_resource_group.main", "azurerm_resource_group"},
			expectedDataSrcs:  []string{},
			expectError:       false,
		},
		{
			name:             "no resources section",
			data:             `## Inputs\n\nSome inputs`,
			providerPrefixes: []string{"azurerm_"},
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMarkdownContent(tt.data, FormatDocument, tt.providerPrefixes)
			resources, dataSources, err := mc.ExtractResourcesAndDataSources()

			if tt.expectError {
				if err == nil {
					t.Error("ExtractResourcesAndDataSources() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractResourcesAndDataSources() unexpected error: %v", err)
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

func TestMatchesSectionName(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
		want     bool
	}{
		{
			name:     "exact match",
			actual:   "Resources",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "case insensitive",
			actual:   "resources",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "plural vs singular",
			actual:   "Resource",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "singular vs plural",
			actual:   "Resources",
			expected: "Resource",
			want:     true,
		},
		{
			name:     "inputs special case",
			actual:   "Required Inputs",
			expected: "Inputs",
			want:     true,
		},
		{
			name:     "optional inputs special case",
			actual:   "Optional Inputs",
			expected: "Inputs",
			want:     true,
		},
		{
			name:     "similar with typo",
			actual:   "Resourses",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "completely different",
			actual:   "Inputs",
			expected: "Resources",
			want:     false,
		},
		{
			name:     "empty strings",
			actual:   "",
			expected: "Resources",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSectionName(tt.actual, tt.expected)
			if result != tt.want {
				t.Errorf("matchesSectionName(%q, %q) = %v; want %v", tt.actual, tt.expected, result, tt.want)
			}
		})
	}
}

func TestMarkdownContent_FallbackSectionItems(t *testing.T) {
	data := `Some content with anchors

<a name="input_var1"></a>
<a name="input_var2"></a>
<a name="output_out1"></a>
`
	mc := NewMarkdownContent(data, FormatDocument, nil)

	tests := []struct {
		name          string
		sections      []string
		expectedItems []string
	}{
		{
			name:          "extract inputs via anchor fallback",
			sections:      []string{"Required Inputs"},
			expectedItems: []string{"var1", "var2"},
		},
		{
			name:          "extract outputs via anchor fallback",
			sections:      []string{"Outputs"},
			expectedItems: []string{"out1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := mc.ExtractSectionItems(tt.sections...)

			if len(items) != len(tt.expectedItems) {
				t.Errorf("ExtractSectionItems() returned %d items; want %d", len(items), len(tt.expectedItems))
				t.Logf("Got: %v", items)
				t.Logf("Want: %v", tt.expectedItems)
			}
		})
	}
}

func TestMarkdownContent_HasProviderPrefix(t *testing.T) {
	tests := []struct {
		name             string
		providerPrefixes []string
		testString       string
		want             bool
	}{
		{
			name:             "matches prefix",
			providerPrefixes: []string{"azurerm_"},
			testString:       "azurerm_resource_group",
			want:             true,
		},
		{
			name:             "no match",
			providerPrefixes: []string{"azurerm_"},
			testString:       "aws_instance",
			want:             false,
		},
		{
			name:             "case insensitive match",
			providerPrefixes: []string{"azurerm_"},
			testString:       "AzureRM_resource_group",
			want:             true,
		},
		{
			name:             "multiple prefixes first matches",
			providerPrefixes: []string{"azurerm_", "aws_"},
			testString:       "azurerm_resource",
			want:             true,
		},
		{
			name:             "multiple prefixes second matches",
			providerPrefixes: []string{"azurerm_", "aws_"},
			testString:       "aws_instance",
			want:             true,
		},
		{
			name:             "empty prefixes",
			providerPrefixes: []string{},
			testString:       "azurerm_resource",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMarkdownContent("", FormatDocument, tt.providerPrefixes)
			result := mc.hasProviderPrefix(tt.testString)
			if result != tt.want {
				t.Errorf("hasProviderPrefix(%q) = %v; want %v", tt.testString, result, tt.want)
			}
		})
	}
}

func TestMarkdownContent_IndexAnchors(t *testing.T) {
	data := `
<a name="input_var1"></a>
<a name="input_var2"></a>
<a name="output_out1"></a>
<a NAME="INPUT_VAR3"></a>
`
	mc := NewMarkdownContent(data, FormatDocument, nil)

	tests := []struct {
		name         string
		anchorName   string
		expectedType string
		shouldExist  bool
	}{
		{
			name:         "input anchor exists",
			anchorName:   "var1",
			expectedType: "input",
			shouldExist:  true,
		},
		{
			name:         "output anchor exists",
			anchorName:   "out1",
			expectedType: "output",
			shouldExist:  true,
		},
		{
			name:         "case insensitive anchor",
			anchorName:   "var3",
			expectedType: "input",
			shouldExist:  true,
		},
		{
			name:        "non-existing anchor",
			anchorName:  "nonexistent",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anchorTypes := mc.anchorTypes[strings.ToLower(tt.anchorName)]
			if tt.shouldExist {
				if len(anchorTypes) == 0 || !anchorTypes[tt.expectedType] {
					t.Errorf("Expected anchor %q with type %q to exist", tt.anchorName, tt.expectedType)
				}
			} else {
				if len(anchorTypes) > 0 {
					t.Errorf("Expected anchor %q not to exist", tt.anchorName)
				}
			}
		})
	}
}

func TestAddUnique(t *testing.T) {
	tests := []struct {
		name          string
		initial       []string
		toAdd         []string
		expectedCount int
		expectedItems []string
	}{
		{
			name:          "add new item",
			initial:       []string{"a", "b"},
			toAdd:         []string{"c"},
			expectedCount: 3,
			expectedItems: []string{"a", "b", "c"},
		},
		{
			name:          "add duplicate item",
			initial:       []string{"a", "b"},
			toAdd:         []string{"a"},
			expectedCount: 2,
			expectedItems: []string{"a", "b"},
		},
		{
			name:          "add multiple items",
			initial:       []string{"a"},
			toAdd:         []string{"b", "c", "a"},
			expectedCount: 3,
			expectedItems: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := tt.initial
			for _, item := range tt.toAdd {
				addUnique(&slice, item)
			}

			if len(slice) != tt.expectedCount {
				t.Errorf("addUnique resulted in %d items; want %d", len(slice), tt.expectedCount)
			}

			for _, expected := range tt.expectedItems {
				if !slices.Contains(slice, expected) {
					t.Errorf("Expected item %q not found in slice", expected)
				}
			}
		})
	}
}
