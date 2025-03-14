package markparsr

import (
	"fmt"
	"strings"
)

// compareTerraformAndMarkdown identifies discrepancies between terraform code and
// markdown documentation. It handles both full resource names (azurerm_resource_group.example)
// and resource types (azurerm_resource_group), considering a match if either form is documented.
func compareTerraformAndMarkdown(tfItems, mdItems []string, itemType string) []error {
	errors := make([]error, 0, len(tfItems)+len(mdItems))
	tfSet := make(map[string]bool, len(tfItems)*2)
	mdSet := make(map[string]bool, len(mdItems)*2)
	reported := make(map[string]bool, len(tfItems)+len(mdItems))

	getFullName := func(items []string, baseName string) string {
		for _, item := range items {
			if strings.HasPrefix(item, baseName+".") {
				return item
			}
		}
		return baseName
	}

	// Add both full names and base types to the sets
	for _, item := range tfItems {
		tfSet[item] = true
		baseName := strings.Split(item, ".")[0]
		tfSet[baseName] = true
	}

	for _, item := range mdItems {
		mdSet[item] = true
		baseName := strings.Split(item, ".")[0]
		mdSet[baseName] = true
	}

	// Find items in terraform but not in markdown
	for _, tfItem := range tfItems {
		baseName := strings.Split(tfItem, ".")[0]
		if !mdSet[tfItem] && !mdSet[baseName] && !reported[baseName] {
			fullName := getFullName(tfItems, baseName)
			errors = append(errors, fmt.Errorf("%s in Terraform but missing in markdown: %s", itemType, fullName))
			reported[baseName] = true
		}
	}

	// Find items in markdown but not in terraform
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
