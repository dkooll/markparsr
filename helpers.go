package markparsr

import (
	"fmt"
	"strings"
)

// compareTerraformAndMarkdown compares items defined in Terraform with those documented in markdown.
// It identifies items that are in Terraform but missing from the documentation and vice versa.
// The function handles both full resource names (e.g., "azurerm_resource_group.example")
// and base resource types (e.g., "azurerm_resource_group").
// Parameters:
//   - tfItems: Slice of items found in Terraform code
//   - mdItems: Slice of items found in markdown documentation
//   - itemType: Description of the type of items (e.g., "Resources", "Data Sources")
//
// Returns:
//   - A slice of errors describing mismatches between Terraform and markdown
func compareTerraformAndMarkdown(tfItems, mdItems []string, itemType string) []error {
	errors := make([]error, 0, len(tfItems)+len(mdItems))
	tfSet := make(map[string]bool, len(tfItems)*2)
	mdSet := make(map[string]bool, len(mdItems)*2)
	reported := make(map[string]bool, len(tfItems)+len(mdItems))

	// getFullName returns the full resource name for a base resource type
	// by finding any item that starts with the base name followed by a period.
	getFullName := func(items []string, baseName string) string {
		for _, item := range items {
			if strings.HasPrefix(item, baseName+".") {
				return item
			}
		}
		return baseName
	}

	// Add both full names and base types to the Terraform set
	for _, item := range tfItems {
		tfSet[item] = true
		baseName := strings.Split(item, ".")[0]
		tfSet[baseName] = true
	}

	// Add both full names and base types to the markdown set
	for _, item := range mdItems {
		mdSet[item] = true
		baseName := strings.Split(item, ".")[0]
		mdSet[baseName] = true
	}

	// Find items in Terraform but not in markdown
	for _, tfItem := range tfItems {
		baseName := strings.Split(tfItem, ".")[0]
		if !mdSet[tfItem] && !mdSet[baseName] && !reported[baseName] {
			fullName := getFullName(tfItems, baseName)
			errors = append(errors, fmt.Errorf("%s in Terraform but missing in markdown: %s", itemType, fullName))
			reported[baseName] = true
		}
	}

	// Find items in markdown but not in Terraform
	for _, mdItem := range mdItems {
		baseName := strings.Split(mdItem, ".")[0]
		if !tfSet[mdItem] && !tfSet[baseName] && !reported[baseName] {
			fullName := getFullName(mdItems, baseName)
			errors = append(errors, fmt.Errorf("%s in markdown but missing in Terraform: %s", itemType, fullName))
			reported[baseName] = true
		}
	}
	return errors
}
