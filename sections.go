package markparsr

import (
	"fmt"
)

type SectionValidator struct {
	content  *MarkdownContent
	sections []string
}

func NewSectionValidator(content *MarkdownContent) *SectionValidator {
	sections := []string{
		"Goals", "Resources", "Providers", "Requirements",
		"Optional Inputs", "Required Inputs", "Outputs", "Testing",
	}
	return &SectionValidator{content: content, sections: sections}
}

func (sv *SectionValidator) Validate() []error {
	var allErrors []error

	for _, section := range sv.sections {
		if !sv.content.HasSection(section) {
			allErrors = append(allErrors, fmt.Errorf("required section missing: '%s'", section))
		}
	}

	return allErrors
}
