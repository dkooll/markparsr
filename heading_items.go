package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// HeadingItemValidator validates items where they are under H3 headings
type HeadingItemValidator struct {
	data      string
	itemType  string
	blockType string
	sections  []string
	fileName  string
}

// NewHeadingItemValidator creates a new HeadingItemValidator
func NewHeadingItemValidator(data, itemType, blockType string, sections []string, fileName string) *HeadingItemValidator {
	return &HeadingItemValidator{
		data:      data,
		itemType:  itemType,
		blockType: blockType,
		sections:  sections,
		fileName:  fileName,
	}
}

// Validate compares Terraform items with those documented in the markdown
func (iv *HeadingItemValidator) Validate() []error {
	workspace := os.Getenv("GITHUB_WORKSPACE")
	if workspace == "" {
		var err error
		workspace, err = os.Getwd()
		if err != nil {
			return []error{fmt.Errorf("failed to get current working directory: %v", err)}
		}
	}
	filePath := filepath.Join(workspace, "caller", iv.fileName)
	tfItems, err := extractTerraformItems(filePath, iv.blockType)
	if err != nil {
		return []error{err}
	}

	var mdItems []string
	for _, section := range iv.sections {
		items := extractHeadingMarkdownSectionItems(iv.data, section)
		mdItems = append(mdItems, items...)
	}

	return compareTerraformAndMarkdown(tfItems, mdItems, iv.itemType)
}

// extractHeadingMarkdownSectionItems extracts items from heading-based markdown sections
func extractHeadingMarkdownSectionItems(data string, sectionName string) []string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	rootNode := p.Parse([]byte(data))

	var items []string
	inTargetSection := false

	ast.WalkFunc(rootNode, func(n ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := n.(*ast.Heading); ok && entering {
			headingText := strings.TrimSpace(extractText(heading))
			if heading.Level == 2 {
				inTargetSection = false
				if strings.EqualFold(headingText, sectionName) {
					inTargetSection = true
				}
			} else if heading.Level == 3 && inTargetSection {
				inputName := strings.Trim(headingText, " []")
				items = append(items, inputName)
			}
		}
		return ast.GoToNext
	})

	return items
}
