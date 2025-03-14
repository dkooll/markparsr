package markparsr

import (
	"errors"
	"slices"
	"strings"
	"sync"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// MarkdownContent parses and analyzes Terraform module documentation.
// It provides methods to check for sections and extract documented resources.
type MarkdownContent struct {
	data       string
	rootNode   ast.Node
	parser     *parser.Parser
	sections   map[string]bool
	stringPool *sync.Pool
}

// NewMarkdownContent creates a new analyzer for markdown content.
// It parses the markdown into an AST for efficient analysis.
func NewMarkdownContent(data string) *MarkdownContent {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	rootNode := markdown.Parse([]byte(data), p)

	return &MarkdownContent{
		data:     data,
		rootNode: rootNode,
		parser:   p,
		sections: make(map[string]bool),
		stringPool: &sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
	}
}

// HasSection checks if a section exists in the markdown.
// Section matching is case-insensitive and handles plural forms.
// For "Inputs", it also matches "Required Inputs" or "Optional Inputs".
func (mc *MarkdownContent) HasSection(sectionName string) bool {
	if found, exists := mc.sections[sectionName]; exists {
		return found
	}

	found := false
	ast.WalkFunc(mc.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering && heading.Level == 2 {
			text := strings.TrimSpace(mc.extractText(heading))
			if strings.EqualFold(text, sectionName) ||
				strings.EqualFold(text, sectionName+"s") ||
				(sectionName == "Inputs" && (strings.EqualFold(text, "Required Inputs") || strings.EqualFold(text, "Optional Inputs"))) {
				found = true
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})

	mc.sections[sectionName] = found
	return found
}

// ExtractSectionItems extracts item names from level 3 headings within specified sections.
// This is typically used to find documented variables and outputs.
func (mc *MarkdownContent) ExtractSectionItems(sectionNames ...string) []string {
	var items []string
	inTargetSection := false

	ast.WalkFunc(mc.rootNode, func(n ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := n.(*ast.Heading); ok && entering {
			headingText := strings.TrimSpace(mc.extractText(heading))
			if heading.Level == 2 {
				inTargetSection = false
				for _, sectionName := range sectionNames {
					if strings.EqualFold(headingText, sectionName) {
						inTargetSection = true
						break
					}
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

// ExtractResourcesAndDataSources finds Terraform resources and data sources in the markdown.
// It looks for links containing "azurerm_" in the Resources section, distinguishing
// between regular resources and data sources based on URL patterns.
func (mc *MarkdownContent) ExtractResourcesAndDataSources() ([]string, []string, error) {
	var resources []string
	var dataSources []string
	inResourceSection := false

	ast.WalkFunc(mc.rootNode, func(n ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := n.(*ast.Heading); ok && entering {
			headingText := mc.extractText(heading)
			if strings.Contains(headingText, "Resources") {
				inResourceSection = true
			} else if heading.Level <= 2 {
				inResourceSection = false
			}
		}
		if inResourceSection && entering {
			if link, ok := n.(*ast.Link); ok {
				linkText := mc.extractText(link)
				destination := string(link.Destination)
				if strings.Contains(linkText, "azurerm_") {
					resourceName := strings.Split(linkText, "]")[0]
					resourceName = strings.TrimPrefix(resourceName, "[")
					baseName := strings.Split(resourceName, ".")[0]

					if strings.Contains(destination, "/data-sources/") {
						if !slices.Contains(dataSources, resourceName) {
							dataSources = append(dataSources, resourceName)
						}
						if !slices.Contains(dataSources, baseName) {
							dataSources = append(dataSources, baseName)
						}
					} else {
						if !slices.Contains(resources, resourceName) {
							resources = append(resources, resourceName)
						}
						if !slices.Contains(resources, baseName) {
							resources = append(resources, baseName)
						}
					}
				}
			}
		}
		return ast.GoToNext
	})

	if len(resources) == 0 && len(dataSources) == 0 {
		return nil, nil, errors.New("resources section not found or empty")
	}

	return resources, dataSources, nil
}

// extractText gets the text content from a node, using a string pool for efficiency.
func (mc *MarkdownContent) extractText(node ast.Node) string {
	sb := mc.stringPool.Get().(*strings.Builder)
	sb.Reset()
	defer mc.stringPool.Put(sb)

	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch tn := n.(type) {
			case *ast.Text:
				sb.Write(tn.Literal)
			case *ast.Code:
				sb.Write(tn.Literal)
			}
		}
		return ast.GoToNext
	})

	return sb.String()
}
