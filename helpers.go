package markparsr

import (
	"fmt"
	"strings"
)

// compareTerraformAndMarkdown identifies discrepancies between terraform code and
// markdown documentation. It reports ALL instances of resources that are missing
// from either side, without any deduplication or special treatment for resource types.
func compareTerraformAndMarkdown(tfItems, mdItems []string, itemType string) []error {
	errors := make([]error, 0)

	// Create maps for exact matching of full resource names
	tfFullNames := make(map[string]bool)
	mdFullNames := make(map[string]bool)

	// Split items into full names (with dots) and base types
	for _, item := range tfItems {
		if strings.Contains(item, ".") {
			tfFullNames[item] = true
		}
	}

	for _, item := range mdItems {
		if strings.Contains(item, ".") {
			mdFullNames[item] = true
		}
	}

	// Find terraform resources missing from markdown
	for tfItem := range tfFullNames {
		if !mdFullNames[tfItem] {
			errors = append(errors, fmt.Errorf("%s in Terraform but missing in markdown: %s", itemType, tfItem))
		}
	}

	// Find markdown resources missing from terraform
	for mdItem := range mdFullNames {
		if !tfFullNames[mdItem] {
			errors = append(errors, fmt.Errorf("%s in markdown but missing in Terraform: %s", itemType, mdItem))
		}
	}

	return errors
}
