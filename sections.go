package markparsr

import (
	"fmt"
	"slices"
	"strings"
)

type SectionValidator struct {
	content            *MarkdownContent
	requiredSections   []string
	additionalSections []string
}

func NewSectionValidator(content *MarkdownContent, additionalSections []string) *SectionValidator {
	requiredSections := []string{
		"Resources", "Providers", "Requirements",
	}

	requiredSections = append(requiredSections, "Required Inputs", "Optional Inputs", "Outputs")

	return &SectionValidator{
		content:            content,
		requiredSections:   requiredSections,
		additionalSections: additionalSections,
	}
}

func (sv *SectionValidator) Validate() []error {
	var allErrors []error
	foundSections := sv.content.GetAllSections()

	handledSections := make(map[string]bool)

	missingSections := make(map[string]bool)

	for _, requiredSection := range sv.requiredSections {
		if slices.Contains(foundSections, requiredSection) {
			handledSections[requiredSection] = true
			continue
		}

		misspellingFound := false
		for _, foundSection := range foundSections {
			if !handledSections[foundSection] && isSimilarSection(foundSection, requiredSection) {
				allErrors = append(allErrors, fmt.Errorf("section '%s' appears to be misspelled (should be '%s')",
					foundSection, requiredSection))
				handledSections[foundSection] = true
				misspellingFound = true
				break
			}
		}

		if !misspellingFound {
			missingSections[requiredSection] = true
			allErrors = append(allErrors, fmt.Errorf("required section missing: '%s'", requiredSection))
		}
	}

	for _, additionalSection := range sv.additionalSections {
		if handledSections[additionalSection] || missingSections[additionalSection] {
			continue
		}

		if slices.Contains(foundSections, additionalSection) {
			handledSections[additionalSection] = true
			continue
		}

		misspellingFound := false
		for _, foundSection := range foundSections {
			if !handledSections[foundSection] && isSimilarSection(foundSection, additionalSection) {
				allErrors = append(allErrors, fmt.Errorf("section '%s' appears to be misspelled (should be '%s')",
					foundSection, additionalSection))
				handledSections[foundSection] = true
				misspellingFound = true
				break
			}
		}

		if !misspellingFound {
			allErrors = append(allErrors, fmt.Errorf("additional section missing: '%s'", additionalSection))
		}
	}
	return allErrors
}

func isSimilarSection(found, expected string) bool {
	if found == expected {
		return true
	}

	if found+"s" == expected || found == expected+"s" {
		return true
	}

	if levenshtein(found, expected) <= 2 {
		return true
	}

	if strings.EqualFold(found, expected) && found != expected {
		return true
	}

	return false
}
