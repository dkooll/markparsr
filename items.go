package markparsr

import (
	"path/filepath"
	"slices"
)

type ItemValidator struct {
	markdown  *MarkdownContent
	terraform *TerraformContent
	itemType  string
	blockType string
	sections  []string
	fileName  string
}

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

func (iv *ItemValidator) Validate() []error {
	sectionExists := slices.ContainsFunc(iv.sections, func(section string) bool {
		return iv.markdown.HasSection(section)
	})

	if !sectionExists {
		return nil
	}

	filePath := filepath.Join(iv.terraform.workspace, "caller", iv.fileName)
	tfItems, err := iv.terraform.ExtractItems(filePath, iv.blockType)
	if err != nil {
		return []error{err}
	}

	var mdItems []string
	for _, section := range iv.sections {
		mdItems = append(mdItems, iv.markdown.ExtractSectionItems(section)...)
	}

	return compareTerraformAndMarkdown(tfItems, mdItems, iv.itemType)
}
