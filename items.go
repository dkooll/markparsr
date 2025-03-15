package markparsr

import (
	"path/filepath"
	"slices"
)

// ItemValidator ensures that items like variables and outputs in Terraform
// are properly documented in markdown, and vice versa.
type ItemValidator struct {
	markdown  *MarkdownContent
	terraform *TerraformContent
	itemType  string
	blockType string
	sections  []string
	fileName  string
}

// NewItemValidator creates a validator to check that items of a specific type
// (like variables or outputs) are properly documented. It links Terraform blocks
// to their expected markdown sections.
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

// Validate checks that all items in Terraform are documented in markdown and vice versa.
// Validation is skipped if none of the specified markdown sections exist.
func (iv *ItemValidator) Validate() []error {
	// Skip validation if none of the relevant sections exist
	sectionExists := slices.ContainsFunc(iv.sections, func(section string) bool {
		return iv.markdown.HasSection(section)
	})

	if !sectionExists {
		return nil
	}

	// Extract items from Terraform file directly in the module directory
	filePath := filepath.Join(iv.terraform.workspace, iv.fileName)
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
