package markparsr

import (
	"path/filepath"
	"slices"
)

// ItemValidator checks that all items of a specific type (like variables or outputs)
// defined in Terraform files are documented in the markdown, and vice versa.
type ItemValidator struct {
	markdown  *MarkdownContent
	terraform *TerraformContent
	itemType  string
	blockType string
	sections  []string
	fileName  string
}

// NewItemValidator creates a new ItemValidator for checking specific types of Terraform items.
// This validator can be used to check that variables, outputs, or other Terraform
// block types are properly documented in specific sections of the markdown.
// Parameters:
//   - markdown: The parsed markdown content
//   - terraform: The Terraform content analyzer
//   - itemType: Description of the items being validated (e.g., "Variables", "Outputs")
//   - blockType: The Terraform block type to check (e.g., "variable", "output")
//   - sections: Markdown sections where these items should be documented
//   - fileName: Name of the Terraform file containing these items (e.g., "variables.tf")
//
// Returns:
//   - A pointer to the initialized ItemValidator
func NewItemValidator(markdown *MarkdownContent, terraform *TerraformContent, itemType, blockType string, sections []string, fileName string) *ItemValidator {
	return &ItemValidator{
		markdown:  markdown,
		terraform: terraform,
		itemType:  itemType,
		blockType: blockType,
		sections:  sections,
		fileName:  fileName,
	}
}

// Validate checks that all items of the specified type are properly documented.
// The validation is only performed if at least one of the specified markdown
// sections exists. If none of the sections exist, validation is skipped.
// Returns:
//   - A slice of errors describing items that are in Terraform but not documented
//     in markdown, or items that are documented but not defined in Terraform.
//     Empty if all items are properly documented.
func (iv *ItemValidator) Validate() []error {
	// Skip validation if none of the relevant sections exist
	sectionExists := slices.ContainsFunc(iv.sections, func(section string) bool {
		return iv.markdown.HasSection(section)
	})

	if !sectionExists {
		return nil
	}

	// Extract items from Terraform file
	filePath := filepath.Join(iv.terraform.workspace, "caller", iv.fileName)
	tfItems, err := iv.terraform.ExtractItems(filePath, iv.blockType)
	if err != nil {
		return []error{err}
	}

	// Extract items from markdown sections
	var mdItems []string
	for _, section := range iv.sections {
		mdItems = append(mdItems, iv.markdown.ExtractSectionItems(section)...)
	}

	// Compare items in Terraform and markdown
	return compareTerraformAndMarkdown(tfItems, mdItems, iv.itemType)
}
