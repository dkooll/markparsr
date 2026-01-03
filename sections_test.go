package markparsr

import (
	"strings"
	"testing"
)

func TestNewSectionValidator(t *testing.T) {
	tests := []struct {
		name               string
		additionalSections []string
		expectedRequired   int
	}{
		{
			name:               "with no additional sections",
			additionalSections: []string{},
			expectedRequired:   6, // Resources, Providers, Requirements, Required Inputs, Optional Inputs, Outputs
		},
		{
			name:               "with additional sections",
			additionalSections: []string{"Examples", "Notes"},
			expectedRequired:   6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMarkdownContent("", FormatDocument, nil)
			sv := NewSectionValidator(mc, tt.additionalSections)

			if sv == nil {
				t.Fatal("NewSectionValidator() returned nil")
			}

			if len(sv.requiredSections) != tt.expectedRequired {
				t.Errorf("NewSectionValidator() required sections = %d; want %d", len(sv.requiredSections), tt.expectedRequired)
			}

			if len(sv.additionalSections) != len(tt.additionalSections) {
				t.Errorf("NewSectionValidator() additional sections = %d; want %d", len(sv.additionalSections), len(tt.additionalSections))
			}
		})
	}
}

func TestSectionValidator_Validate(t *testing.T) {
	tests := []struct {
		name               string
		markdown           string
		additionalSections []string
		expectedErrors     int
		errorContains      []string
	}{
		{
			name: "all required sections present",
			markdown: `# Test Module

## Resources

content

## Providers

content

## Requirements

content

## Required Inputs

content

## Optional Inputs

content

## Outputs

content
`,
			additionalSections: []string{},
			expectedErrors:     0,
		},
		{
			name: "missing required section",
			markdown: `# Test Module

## Providers

content

## Requirements

content

## Required Inputs

content

## Optional Inputs

content

## Outputs

content
`,
			additionalSections: []string{},
			expectedErrors:     1,
			errorContains:      []string{"Resources", "missing"},
		},
		{
			name: "misspelled section",
			markdown: `# Test Module

## Resourses

content

## Providers

content

## Requirements

content

## Required Inputs

content

## Optional Inputs

content

## Outputs

content
`,
			additionalSections: []string{},
			expectedErrors:     1,
			errorContains:      []string{"Resourses", "misspelled", "Resources"},
		},
		{
			name: "missing additional section",
			markdown: `# Test Module

## Resources

content

## Providers

content

## Requirements

content

## Required Inputs

content

## Optional Inputs

content

## Outputs

content
`,
			additionalSections: []string{"Examples"},
			expectedErrors:     1,
			errorContains:      []string{"Examples", "missing"},
		},
		{
			name: "multiple missing sections",
			markdown: `# Test Module

## Requirements

content
`,
			additionalSections: []string{},
			expectedErrors:     5, // Missing: Resources, Providers, Required Inputs, Optional Inputs, Outputs
			errorContains:      []string{"missing"},
		},
		{
			name: "case variation handled",
			markdown: `# Test Module

## resources

content

## Providers

content

## Requirements

content

## Required Inputs

content

## Optional Inputs

content

## Outputs

content
`,
			additionalSections: []string{},
			expectedErrors:     1,
			errorContains:      []string{"resources", "misspelled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMarkdownContent(tt.markdown, FormatDocument, nil)
			sv := NewSectionValidator(mc, tt.additionalSections)

			errs := sv.Validate()

			if len(errs) != tt.expectedErrors {
				t.Errorf("Validate() returned %d errors; want %d", len(errs), tt.expectedErrors)
				for i, err := range errs {
					t.Logf("  error %d: %v", i+1, err)
				}
			}

			for _, substr := range tt.errorContains {
				found := false
				for _, err := range errs {
					if err != nil && strings.Contains(err.Error(), substr) {
						found = true
						break
					}
				}
				if !found && tt.expectedErrors > 0 {
					t.Errorf("Expected error containing %q, but not found", substr)
				}
			}
		})
	}
}

func TestIsSimilarSection(t *testing.T) {
	tests := []struct {
		name     string
		found    string
		expected string
		want     bool
	}{
		{
			name:     "exact match",
			found:    "Resources",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "plural singular",
			found:    "Resource",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "singular plural",
			found:    "Resources",
			expected: "Resource",
			want:     true,
		},
		{
			name:     "small levenshtein distance",
			found:    "Resourses",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "case difference only",
			found:    "resources",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "completely different",
			found:    "Inputs",
			expected: "Resources",
			want:     false,
		},
		{
			name:     "large levenshtein distance",
			found:    "Something",
			expected: "Resources",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSimilarSection(tt.found, tt.expected)
			if result != tt.want {
				t.Errorf("isSimilarSection(%q, %q) = %v; want %v", tt.found, tt.expected, result, tt.want)
			}
		})
	}
}

func TestSectionValidator_DuplicateHandling(t *testing.T) {
	markdown := `# Test

## Resources

content

## Resources

more content

## Providers

content

## Requirements

content

## Required Inputs

content

## Optional Inputs

content

## Outputs

content
`
	mc := NewMarkdownContent(markdown, FormatDocument, nil)
	sv := NewSectionValidator(mc, []string{})

	errs := sv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with duplicate section returned %d errors; want 0", len(errs))
	}
}

func TestSectionValidator_EmptyMarkdown(t *testing.T) {
	mc := NewMarkdownContent("", FormatDocument, nil)
	sv := NewSectionValidator(mc, []string{})

	errs := sv.Validate()

	if len(errs) != 6 {
		t.Errorf("Validate() on empty markdown returned %d errors; want 6", len(errs))
	}

	for _, err := range errs {
		if !strings.Contains(err.Error(), "missing") {
			t.Errorf("Expected all errors to contain 'missing', got: %v", err)
		}
	}
}

func TestSectionValidator_AdditionalSectionMisspelled(t *testing.T) {
	markdown := `# Test

## Resources

content

## Providers

content

## Requirements

content

## Required Inputs

content

## Optional Inputs

content

## Outputs

content

## Examlpes

content with typo
`
	mc := NewMarkdownContent(markdown, FormatDocument, nil)
	sv := NewSectionValidator(mc, []string{"Examples"})

	errs := sv.Validate()

	if len(errs) != 1 {
		t.Errorf("Validate() returned %d errors; want 1", len(errs))
	}

	if len(errs) > 0 && !strings.Contains(errs[0].Error(), "misspelled") {
		t.Errorf("Expected error about misspelling, got: %v", errs[0])
	}
}

func TestSectionValidator_MixedRequiredAndAdditional(t *testing.T) {
	markdown := `# Test

## Resources

content

## Providers

content

## Requirements

content

## Required Inputs

content

## Optional Inputs

content

## Examples

additional section

## Notes

another additional
`
	mc := NewMarkdownContent(markdown, FormatDocument, nil)
	sv := NewSectionValidator(mc, []string{"Examples", "Notes"})

	errs := sv.Validate()

	if len(errs) != 1 {
		t.Errorf("Validate() returned %d errors; want 1 (missing Outputs)", len(errs))
		for _, err := range errs {
			t.Logf("  error: %v", err)
		}
	}
}

func TestLevenshteinInSections(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "section name with one typo",
			s1:       "Resources",
			s2:       "Resourses",
			expected: 1,
		},
		{
			name:     "section name with two typos",
			s1:       "Providers",
			s2:       "Provuders",
			expected: 1,
		},
		{
			name:     "identical",
			s1:       "Requirements",
			s2:       "Requirements",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := levenshtein(tt.s1, tt.s2)
			if distance != tt.expected {
				t.Errorf("levenshtein(%q, %q) = %d; want %d", tt.s1, tt.s2, distance, tt.expected)
			}
		})
	}
}
