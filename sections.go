package markparsr

import (
	"fmt"
)

// SectionValidator checks for the presence of required sections in markdown documentation.
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
// Each missing section results in an error added to the returned slice.
// Returns:
//   - A slice of errors for missing sections. Empty if all required sections exist.
func (sv *SectionValidator) Validate() []error {
	var allErrors []error

	for _, section := range sv.sections {
		if !sv.content.HasSection(section) {
			allErrors = append(allErrors, fmt.Errorf("required section missing: '%s'", section))
		}
	}

	return allErrors
}
