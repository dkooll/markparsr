package markparsr

import (
	"fmt"
	"slices"
	"strings"
)

// SectionValidator checks for required sections in Terraform module documentation.
type SectionValidator struct {
	content            *MarkdownContent
	requiredSections   []string
	additionalSections []string // Additional sections that should exist and be validated
}

// NewSectionValidator creates a validator for standard Terraform document sections
// additionalSections specifies additional sections that should exist and be validated
func NewSectionValidator(content *MarkdownContent, additionalSections []string) *SectionValidator {
	// Required sections for both formats
	requiredSections := []string{
		"Resources", "Providers", "Requirements",
	}

	// Format-specific required sections
	if content.format == FormatTable {
		requiredSections = append(requiredSections, "Inputs", "Outputs")
	} else {
		// Document format typically has separate sections for required and optional inputs
		requiredSections = append(requiredSections, "Required Inputs", "Optional Inputs", "Outputs")
	}

	return &SectionValidator{
		content:            content,
		requiredSections:   requiredSections,
		additionalSections: additionalSections,
	}
}

// Validate checks that all required and additional sections are present and spelled correctly
func (sv *SectionValidator) Validate() []error {
	var allErrors []error
	foundSections := sv.content.GetAllSections()

	// Track which sections have been handled already
	handledSections := make(map[string]bool)

	// Track sections that are required but missing
	missingSections := make(map[string]bool)

	// First check all required sections
	for _, requiredSection := range sv.requiredSections {
		// Check if the required section exists exactly
		if slices.Contains(foundSections, requiredSection) {
			handledSections[requiredSection] = true
			continue
		}

		// If not found exactly, check for misspellings
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

		// If no misspelling found, mark as missing
		if !misspellingFound {
			missingSections[requiredSection] = true
			allErrors = append(allErrors, fmt.Errorf("required section missing: '%s'", requiredSection))
		}
	}

	// Now check all additional sections - these should exist too
	for _, additionalSection := range sv.additionalSections {
		// Skip if already processed as part of required sections
		if handledSections[additionalSection] || missingSections[additionalSection] {
			continue
		}

		// Check if the additional section exists exactly
		if slices.Contains(foundSections, additionalSection) {
			handledSections[additionalSection] = true
			continue
		}

		// If not found exactly, check for misspellings
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

		// If no misspelling found, mark as missing
		if !misspellingFound {
			allErrors = append(allErrors, fmt.Errorf("additional section missing: '%s'", additionalSection))
		}
	}

	// Any remaining found sections that haven't been handled are ignored
	// (they can exist in any form and won't cause validation errors)

	// For table format, also validate table columns
	if sv.content.format == FormatTable {
		tableErrors := sv.content.ValidateTableColumns()
		allErrors = append(allErrors, tableErrors...)
	}

	return allErrors
}

// isSimilarSection checks if a found section name is likely a typo of an expected section
func isSimilarSection(found, expected string) bool {
	// Exact match
	if found == expected {
		return true
	}

	// Common mistakes: extra 's', missing 's', etc.
	if found+"s" == expected || found == expected+"s" {
		return true
	}

	// Use the levenshtein function from markdown.go
	if levenshtein(found, expected) <= 2 {
		return true
	}

	// Handle case-insensitive matches differently to avoid false positives
	if strings.EqualFold(found, expected) && found != expected {
		return true
	}

	return false
}
