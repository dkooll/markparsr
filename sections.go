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
	additionalSections []string // Optional but validated if present
}

// NewSectionValidator creates a validator for standard Terraform document sections
func NewSectionValidator(content *MarkdownContent) *SectionValidator {
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

	// Additional sections that should be validated if present but aren't required
	additionalSections := []string{
		"Goals", "Testing", "Features", "License", "Authors", "Notes", "Contributing",
		"References", "Non-Goals",
	}

	return &SectionValidator{
		content:            content,
		requiredSections:   requiredSections,
		additionalSections: additionalSections,
	}
}

// Validate checks that all required sections are present and spelled correctly
func (sv *SectionValidator) Validate() []error {
	var allErrors []error
	foundSections := sv.content.GetAllSections()

	// Track which sections have been handled already
	handledSections := make(map[string]bool)
	missingRequired := make(map[string]bool)

	// Map of likely misspellings to correct names
	misspellings := make(map[string]string)

	// First identify misspellings of required sections
	for _, foundSection := range foundSections {
		// Skip if already processed
		if handledSections[foundSection] {
			continue
		}

		// Check if this section is a misspelling of a required section
		for _, requiredSection := range sv.requiredSections {
			if foundSection != requiredSection && isSimilarSection(foundSection, requiredSection) {
				misspellings[foundSection] = requiredSection
				missingRequired[requiredSection] = true
				handledSections[foundSection] = true
				allErrors = append(allErrors, fmt.Errorf("section '%s' appears to be misspelled (should be '%s')",
					foundSection, requiredSection))
				break
			}
		}
	}

	// Check for missing required sections
	for _, requiredSection := range sv.requiredSections {
		// Skip if already marked as missing due to a misspelling
		if missingRequired[requiredSection] {
			continue
		}

		if !slices.Contains(foundSections, requiredSection) {
			allErrors = append(allErrors, fmt.Errorf("required section missing: '%s'", requiredSection))
		} else {
			handledSections[requiredSection] = true
		}
	}

	// Check for other misspellings or unexpected sections
	for _, foundSection := range foundSections {
		// Skip if already handled
		if handledSections[foundSection] {
			continue
		}

		// Check if this is a known additional section
		if slices.Contains(sv.additionalSections, foundSection) {
			handledSections[foundSection] = true
			continue
		}

		// Check if this is a misspelling of an additional section
		misspellingFound := false
		for _, additionalSection := range sv.additionalSections {
			if isSimilarSection(foundSection, additionalSection) {
				allErrors = append(allErrors, fmt.Errorf("section '%s' appears to be misspelled (should be '%s')",
					foundSection, additionalSection))
				handledSections[foundSection] = true
				misspellingFound = true
				break
			}
		}

		// If not a misspelling of anything known, report as unexpected
		if !misspellingFound {
			allErrors = append(allErrors, fmt.Errorf("unexpected section: '%s'", foundSection))
			handledSections[foundSection] = true
		}
	}

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
