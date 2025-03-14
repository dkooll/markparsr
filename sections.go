package markparsr

import (
	"fmt"
)

// SectionValidator checks for required sections in Terraform module documentation.
type SectionValidator struct {
	content  *MarkdownContent
	sections []string
}

// NewSectionValidator creates a new SectionValidator that checks for required sections
func NewSectionValidator(content *MarkdownContent) *SectionValidator {
	sections := []string{
		"Goals", "Resources", "Providers", "Requirements",
		"Optional Inputs", "Required Inputs", "Outputs", "Testing",
	}
	return &SectionValidator{content: content, sections: sections}
}

// Validate checks that all required sections are present in the markdown.
// Reports an error for each missing section.
func (sv *SectionValidator) Validate() []error {
	var allErrors []error

	for _, section := range sv.sections {
		if !sv.content.HasSection(section) {
			allErrors = append(allErrors, fmt.Errorf("required section missing: '%s'", section))
		}
	}

	return allErrors
}
