package markparsr

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// Section represents a section in the markdown with required table columns
type Section struct {
	Header       string
	RequiredCols []string
	OptionalCols []string
}

// SectionValidator validates markdown sections
type SectionValidator struct {
	data     string
	sections []Section
	rootNode ast.Node
}

// NewSectionValidator creates a new SectionValidator
func NewSectionValidator(data string) *SectionValidator {
	sections := []Section{
		{Header: "Goals"},
		{Header: "Non-Goals"},
		{Header: "Resources", RequiredCols: []string{"Name", "Type"}},
		{Header: "Providers", RequiredCols: []string{"Name", "Version"}},
		{Header: "Requirements", RequiredCols: []string{"Name", "Version"}},
		{Header: "Inputs",
			RequiredCols: []string{"Name", "Description", "Required"},
			OptionalCols: []string{"Type", "Default"},
		},
		{Header: "Outputs", RequiredCols: []string{"Name", "Description"}},
		{Header: "Features"},
		{Header: "Testing"},
		{Header: "Authors"},
		{Header: "License"},
		{Header: "Notes"},
		{Header: "Contributing"},
		{Header: "References"},
	}

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	rootNode := markdown.Parse([]byte(data), p)

	return &SectionValidator{
		data:     data,
		sections: sections,
		rootNode: rootNode,
	}
}

// Validate validates the sections in the markdown
func (sv *SectionValidator) Validate() []error {
	var allErrors []error
	for _, section := range sv.sections {
		allErrors = append(allErrors, section.validate(sv.rootNode)...)
	}
	return allErrors
}

// validate checks if a section and its columns are correctly formatted
func (s Section) validate(rootNode ast.Node) []error {
	var errors []error
	found := false

	ast.WalkFunc(rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering && heading.Level == 2 {
			text := strings.TrimSpace(extractText(heading))
			if text == s.Header { // exact match
				found = true
				if len(s.RequiredCols) > 0 || len(s.OptionalCols) > 0 {
					nextNode := getNextSibling(node)
					if table, ok := nextNode.(*ast.Table); ok {
						actualHeaders, err := extractTableHeaders(table)
						if err != nil {
							errors = append(errors, err)
						} else {
							errors = append(errors, validateColumns(s.Header, s.RequiredCols, s.OptionalCols, actualHeaders)...)
						}
					} else {
						errors = append(errors, formatError("missing table after header: %s", s.Header))
					}
				}
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})

	if !found {
		errors = append(errors, compareHeaders(s.Header, ""))
	}

	return errors
}
