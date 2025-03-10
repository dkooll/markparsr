package markparsr

import (
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// HeadingSectionValidator validates sections where items are under headings, not in tables
type HeadingSectionValidator struct {
	data     string
	sections []string
	rootNode ast.Node
}

// NewHeadingSectionValidator creates a new HeadingSectionValidator
func NewHeadingSectionValidator(data string) *HeadingSectionValidator {
	sections := []string{
		"Goals", "Resources", "Providers", "Requirements",
		"Optional Inputs", "Required Inputs", "Outputs", "Testing",
	}

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	rootNode := markdown.Parse([]byte(data), p)

	return &HeadingSectionValidator{data: data, sections: sections, rootNode: rootNode}
}

// Validate validates the sections in the markdown
func (sv *HeadingSectionValidator) Validate() []error {
	var allErrors []error

	// Check each required section
	for _, section := range sv.sections {
		if !sv.validateSection(section) {
			// Add explicit error for missing section
			allErrors = append(allErrors, fmt.Errorf("required section missing: '%s'", section))
		}
	}

	return allErrors
}

// validateSection checks if a section exists in the markdown
func (sv *HeadingSectionValidator) validateSection(sectionName string) bool {
	found := false
	ast.WalkFunc(sv.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering && heading.Level == 2 {
			text := strings.TrimSpace(extractText(heading))
			if strings.EqualFold(text, sectionName) ||
				strings.EqualFold(text, sectionName+"s") ||
				(sectionName == "Inputs" && (strings.EqualFold(text, "Required Inputs") || strings.EqualFold(text, "Optional Inputs"))) {
				found = true
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})
	return found
}
