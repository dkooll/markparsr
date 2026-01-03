package markparsr

import "testing"

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "identical strings",
			s1:       "hello",
			s2:       "hello",
			expected: 0,
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "",
			expected: 0,
		},
		{
			name:     "first string empty",
			s1:       "",
			s2:       "hello",
			expected: 5,
		},
		{
			name:     "second string empty",
			s1:       "hello",
			s2:       "",
			expected: 5,
		},
		{
			name:     "single character difference",
			s1:       "hello",
			s2:       "hallo",
			expected: 1,
		},
		{
			name:     "multiple differences",
			s1:       "kitten",
			s2:       "sitting",
			expected: 3,
		},
		{
			name:     "section name typo",
			s1:       "Resources",
			s2:       "Resourses",
			expected: 1,
		},
		{
			name:     "completely different",
			s1:       "abc",
			s2:       "xyz",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := levenshtein(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("levenshtein(%q, %q) = %d; want %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

func TestDefaultStringUtils_LevenshteinDistance(t *testing.T) {
	utils := NewStringUtils()

	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "identical strings",
			s1:       "test",
			s2:       "test",
			expected: 0,
		},
		{
			name:     "one character difference",
			s1:       "test",
			s2:       "best",
			expected: 1,
		},
		{
			name:     "empty first string",
			s1:       "",
			s2:       "test",
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.LevenshteinDistance(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

func TestDefaultStringUtils_IsSimilarSection(t *testing.T) {
	utils := NewStringUtils()

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
			name:     "case difference",
			found:    "resources",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "plural singular match",
			found:    "Resource",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "singular plural match",
			found:    "Resources",
			expected: "Resource",
			want:     true,
		},
		{
			name:     "small typo within threshold",
			found:    "Resourses",
			expected: "Resources",
			want:     true,
		},
		{
			name:     "completely different",
			found:    "Inputs",
			expected: "Resources",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsSimilarSection(tt.found, tt.expected)
			if result != tt.want {
				t.Errorf("IsSimilarSection(%q, %q) = %v; want %v", tt.found, tt.expected, result, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b, c  int
		expected int
	}{
		{
			name:     "a is minimum",
			a:        1,
			b:        2,
			c:        3,
			expected: 1,
		},
		{
			name:     "b is minimum",
			a:        3,
			b:        1,
			c:        2,
			expected: 1,
		},
		{
			name:     "c is minimum",
			a:        3,
			b:        2,
			c:        1,
			expected: 1,
		},
		{
			name:     "all equal",
			a:        5,
			b:        5,
			c:        5,
			expected: 5,
		},
		{
			name:     "negative numbers",
			a:        -1,
			b:        0,
			c:        1,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b, tt.c)
			if result != tt.expected {
				t.Errorf("min(%d, %d, %d) = %d; want %d", tt.a, tt.b, tt.c, result, tt.expected)
			}
		})
	}
}
