package markparsr

import (
	"fmt"
	"testing"
)

func TestBuildItemIndex(t *testing.T) {
	tests := []struct {
		name              string
		items             []string
		expectedEntries   int
		expectedBaseCount map[string]int
	}{
		{
			name:              "empty items",
			items:             []string{},
			expectedEntries:   0,
			expectedBaseCount: map[string]int{},
		},
		{
			name:            "simple items without dots",
			items:           []string{"foo", "bar", "baz"},
			expectedEntries: 3,
			expectedBaseCount: map[string]int{
				"foo": 1,
				"bar": 1,
				"baz": 1,
			},
		},
		{
			name:            "items with qualified names",
			items:           []string{"azurerm_resource.name", "azurerm_data.name"},
			expectedEntries: 2,
			expectedBaseCount: map[string]int{
				"azurerm_resource": 1,
				"azurerm_data":     1,
			},
		},
		{
			name:            "mixed qualified and base names filters base when qualified exists",
			items:           []string{"azurerm_resource", "azurerm_resource.name"},
			expectedEntries: 1,
			expectedBaseCount: map[string]int{
				"azurerm_resource": 1,
			},
		},
		{
			name:            "duplicate items",
			items:           []string{"foo", "foo", "bar"},
			expectedEntries: 2,
			expectedBaseCount: map[string]int{
				"foo": 1,
				"bar": 1,
			},
		},
		{
			name:            "items with whitespace",
			items:           []string{"  foo  ", "bar", " baz "},
			expectedEntries: 3,
			expectedBaseCount: map[string]int{
				"foo": 1,
				"bar": 1,
				"baz": 1,
			},
		},
		{
			name:            "empty strings ignored",
			items:           []string{"", "  ", "foo"},
			expectedEntries: 1,
			expectedBaseCount: map[string]int{
				"foo": 1,
			},
		},
		{
			name:            "case insensitive",
			items:           []string{"Foo", "FOO", "foo"},
			expectedEntries: 1,
			expectedBaseCount: map[string]int{
				"foo": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := buildItemIndex(tt.items)

			if len(index.entries) != tt.expectedEntries {
				t.Errorf("buildItemIndex() entries count = %d; want %d", len(index.entries), tt.expectedEntries)
			}

			for base, expectedCount := range tt.expectedBaseCount {
				actualCount := len(index.byBase[base])
				if actualCount != expectedCount {
					t.Errorf("buildItemIndex() base %q count = %d; want %d", base, actualCount, expectedCount)
				}
			}
		})
	}
}

func TestItemIndex_Items(t *testing.T) {
	t.Run("returns all normalized items", func(t *testing.T) {
		index := buildItemIndex([]string{"foo", "bar", "baz"})
		items := index.items()

		if len(items) != 3 {
			t.Errorf("index.items() count = %d; want 3", len(items))
		}
	})

	t.Run("returns empty slice for empty index", func(t *testing.T) {
		index := buildItemIndex([]string{})
		items := index.items()

		if items == nil {
			t.Error("index.items() returned nil; want empty slice")
		}
		if len(items) != 0 {
			t.Errorf("index.items() count = %d; want 0", len(items))
		}
	})
}

func TestItemIndex_HasMatch(t *testing.T) {
	tests := []struct {
		name        string
		indexItems  []string
		target      string
		shouldMatch bool
	}{
		{
			name:        "exact match",
			indexItems:  []string{"foo", "bar"},
			target:      "foo",
			shouldMatch: true,
		},
		{
			name:        "no match",
			indexItems:  []string{"foo", "bar"},
			target:      "baz",
			shouldMatch: false,
		},
		{
			name:        "base name match",
			indexItems:  []string{"azurerm_resource.name"},
			target:      "azurerm_resource",
			shouldMatch: true,
		},
		{
			name:        "qualified name matches base",
			indexItems:  []string{"azurerm_resource"},
			target:      "azurerm_resource.something",
			shouldMatch: true,
		},
		{
			name:        "case insensitive match",
			indexItems:  []string{"Foo"},
			target:      "foo",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := buildItemIndex(tt.indexItems)
			targetIndex := buildItemIndex([]string{tt.target})
			targetItem := targetIndex.items()[0]

			hasMatch := index.hasMatch(targetItem)
			if hasMatch != tt.shouldMatch {
				t.Errorf("index.hasMatch(%q) = %v; want %v", tt.target, hasMatch, tt.shouldMatch)
			}
		})
	}
}

func TestCompareTerraformAndMarkdown(t *testing.T) {
	tests := []struct {
		name          string
		tfItems       []string
		mdItems       []string
		itemType      string
		expectedErrs  int
		shouldContain []string
	}{
		{
			name:         "all items match",
			tfItems:      []string{"foo", "bar"},
			mdItems:      []string{"foo", "bar"},
			itemType:     "Variables",
			expectedErrs: 0,
		},
		{
			name:          "item in terraform but missing in markdown",
			tfItems:       []string{"foo", "bar"},
			mdItems:       []string{"foo"},
			itemType:      "Variables",
			expectedErrs:  1,
			shouldContain: []string{"bar", "missing in markdown"},
		},
		{
			name:          "item in markdown but missing in terraform",
			tfItems:       []string{"foo"},
			mdItems:       []string{"foo", "bar"},
			itemType:      "Outputs",
			expectedErrs:  1,
			shouldContain: []string{"bar", "missing in Terraform"},
		},
		{
			name:          "multiple mismatches",
			tfItems:       []string{"foo", "bar"},
			mdItems:       []string{"baz", "qux"},
			itemType:      "Resources",
			expectedErrs:  4,
			shouldContain: []string{"foo", "bar", "baz", "qux"},
		},
		{
			name:         "empty lists",
			tfItems:      []string{},
			mdItems:      []string{},
			itemType:     "Variables",
			expectedErrs: 0,
		},
		{
			name:         "qualified names match",
			tfItems:      []string{"azurerm_resource.name"},
			mdItems:      []string{"azurerm_resource.name"},
			itemType:     "Resources",
			expectedErrs: 0,
		},
		{
			name:         "base name matches qualified",
			tfItems:      []string{"azurerm_resource.name"},
			mdItems:      []string{"azurerm_resource"},
			itemType:     "Resources",
			expectedErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := compareTerraformAndMarkdown(tt.tfItems, tt.mdItems, tt.itemType)

			if len(errs) != tt.expectedErrs {
				t.Errorf("compareTerraformAndMarkdown() returned %d errors; want %d", len(errs), tt.expectedErrs)
				for _, err := range errs {
					t.Logf("  error: %v", err)
				}
			}

			for _, substr := range tt.shouldContain {
				found := false
				for _, err := range errs {
					if err != nil && containsString(err.Error(), substr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, but not found", substr)
				}
			}
		})
	}
}

func TestDefaultComparisonValidator_ValidateItems(t *testing.T) {
	validator := NewComparisonValidator()

	tests := []struct {
		name         string
		tfItems      []string
		mdItems      []string
		itemType     string
		expectedErrs int
	}{
		{
			name:         "matching items",
			tfItems:      []string{"var1", "var2"},
			mdItems:      []string{"var1", "var2"},
			itemType:     "Variables",
			expectedErrs: 0,
		},
		{
			name:         "mismatched items",
			tfItems:      []string{"var1"},
			mdItems:      []string{"var2"},
			itemType:     "Outputs",
			expectedErrs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateItems(tt.tfItems, tt.mdItems, tt.itemType)
			if len(errs) != tt.expectedErrs {
				t.Errorf("ValidateItems() returned %d errors; want %d", len(errs), tt.expectedErrs)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr ||
			(len(s) >= len(substr) && findSubstring(s, substr)))))
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNormalizedItem(t *testing.T) {
	tests := []struct {
		name         string
		items        []string
		checkItem    string
		expectedKey  string
		expectedBase string
		expectedDot  bool
	}{
		{
			name:         "simple item",
			items:        []string{"foo"},
			checkItem:    "foo",
			expectedKey:  "foo",
			expectedBase: "foo",
			expectedDot:  false,
		},
		{
			name:         "qualified item",
			items:        []string{"azurerm_resource.name"},
			checkItem:    "azurerm_resource.name",
			expectedKey:  "azurerm_resource.name",
			expectedBase: "azurerm_resource",
			expectedDot:  true,
		},
		{
			name:         "item with whitespace",
			items:        []string{"  foo  "},
			checkItem:    "foo",
			expectedKey:  "foo",
			expectedBase: "foo",
			expectedDot:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := buildItemIndex(tt.items)
			if item, ok := index.entries[tt.expectedKey]; ok {
				if item.key != tt.expectedKey {
					t.Errorf("normalizedItem.key = %q; want %q", item.key, tt.expectedKey)
				}
				if item.base != tt.expectedBase {
					t.Errorf("normalizedItem.base = %q; want %q", item.base, tt.expectedBase)
				}
				if item.hasDot != tt.expectedDot {
					t.Errorf("normalizedItem.hasDot = %v; want %v", item.hasDot, tt.expectedDot)
				}
			} else {
				t.Errorf("Expected to find item with key %q in index", tt.expectedKey)
			}
		})
	}
}

func ExampleNewComparisonValidator() {
	tfItems := []string{"var1", "var2"}
	mdItems := []string{"var1", "var3"}

	validator := NewComparisonValidator()
	errs := validator.ValidateItems(tfItems, mdItems, "Variables")

	for _, err := range errs {
		fmt.Println(err)
	}
}
