package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ItemValidator validates items in Terraform and markdown
type ItemValidator struct {
	data      string
	itemType  string
	blockType string
	section   string
	fileName  string
}

// NewItemValidator creates a new ItemValidator
func NewItemValidator(data, itemType, blockType, section, fileName string) *ItemValidator {
	return &ItemValidator{
		data:      data,
		itemType:  itemType,
		blockType: blockType,
		section:   section,
		fileName:  fileName,
	}
}

// Validate compares Terraform items with those documented in the markdown
func (iv *ItemValidator) Validate() []error {
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

	mdItems, err := extractMarkdownSectionItems(iv.data, iv.section)
	if err != nil {
		return []error{err}
	}

	return compareTerraformAndMarkdown(tfItems, mdItems, iv.itemType)
}

// extractTerraformItems extracts item names from a Terraform file given the block type
func extractTerraformItems(filePath string, blockType string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filepath.Base(filePath), err)
	}

	parser := hclparse.NewParser()
	file, parseDiags := parser.ParseHCL(content, filePath)
	if parseDiags.HasErrors() {
		return nil, fmt.Errorf("error parsing HCL in %s: %v", filepath.Base(filePath), parseDiags)
	}

	var items []string
	body := file.Body

	var diags hcl.Diagnostics

	hclContent, _, contentDiags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: blockType, LabelNames: []string{"name"}},
		},
	})

	diags = append(diags, contentDiags...)

	diags = filterUnsupportedBlockDiagnostics(diags)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error getting content from %s: %v", filepath.Base(filePath), diags)
	}

	if hclContent == nil {
		return items, nil
	}

	for _, block := range hclContent.Blocks {
		if len(block.Labels) > 0 {
			itemName := strings.TrimSpace(block.Labels[0])
			items = append(items, itemName)
		}
	}

	return items, nil
}

// extractMarkdownSectionItems extracts items from a markdown section
func extractMarkdownSectionItems(data, sectionName string) ([]string, error) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	rootNode := markdown.Parse([]byte(data), p)

	var items []string
	var inTargetSection bool

	ast.WalkFunc(rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering && heading.Level == 2 {
			text := strings.TrimSpace(extractText(heading))
			if strings.EqualFold(text, sectionName) || strings.EqualFold(text, sectionName+"s") {
				inTargetSection = true
				return ast.GoToNext
			}
			inTargetSection = false
		}

		if inTargetSection {
			if table, ok := node.(*ast.Table); ok && entering {
				// Extract items from the table
				var bodyNode *ast.TableBody
				for _, child := range table.GetChildren() {
					if body, ok := child.(*ast.TableBody); ok {
						bodyNode = body
						break
					}
				}
				if bodyNode == nil {
					return ast.GoToNext
				}

				for _, rowChild := range bodyNode.GetChildren() {
					if tableRow, ok := rowChild.(*ast.TableRow); ok {
						cells := tableRow.GetChildren()
						if len(cells) > 0 {
							if cell, ok := cells[0].(*ast.TableCell); ok {
								item := extractTextFromNodes(cell.GetChildren())
								item = strings.TrimSpace(item)
								item = strings.Trim(item, "`") // Remove backticks if present
								item = strings.TrimSpace(item)
								items = append(items, item)
							}
						}
					}
				}
				inTargetSection = false
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})

	if len(items) == 0 {
		return nil, fmt.Errorf("%s section not found or empty", sectionName)
	}

	return items, nil
}

// compareTerraformAndMarkdown compares items in Terraform and markdown
func compareTerraformAndMarkdown(tfItems, mdItems []string, itemType string) []error {
	var errors []error

	missingInMarkdown := findMissingItems(tfItems, mdItems)
	if len(missingInMarkdown) > 0 {
		errors = append(errors, formatError("%s missing in markdown:\n  %s", itemType, strings.Join(missingInMarkdown, "\n  ")))
	}

	missingInTerraform := findMissingItems(mdItems, tfItems)
	if len(missingInTerraform) > 0 {
		errors = append(errors, formatError("%s in markdown but missing in Terraform:\n  %s", itemType, strings.Join(missingInTerraform, "\n  ")))
	}

	return errors
}
