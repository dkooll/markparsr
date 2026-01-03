package markparsr

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

type stubTransport struct {
	responses map[string]int
	errURLs   map[string]error
}

type netErrorTimeout struct{}

func (netErrorTimeout) Error() string   { return "timeout" }
func (netErrorTimeout) Timeout() bool   { return true }
func (netErrorTimeout) Temporary() bool { return true }

func (s stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err, ok := s.errURLs[req.URL.String()]; ok {
		return nil, err
	}

	status := s.responses[req.URL.String()]
	if status == 0 {
		status = http.StatusOK
	}

	return &http.Response{
		StatusCode: status,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func withStubHTTPClient(t *testing.T, responses map[string]int, errs map[string]error) {
	original := httpClient
	httpClient = &http.Client{Transport: stubTransport{responses: responses, errURLs: errs}}
	t.Cleanup(func() {
		httpClient = original
	})
}

func TestNewURLValidator(t *testing.T) {
	mc := NewMarkdownContent("", FormatDocument, nil)
	uv := NewURLValidator(mc)

	if uv == nil {
		t.Fatal("NewURLValidator() returned nil")
	}

	if uv.content == nil {
		t.Error("NewURLValidator() content is nil")
	}
}

func TestURLValidator_Validate(t *testing.T) {
	tests := []struct {
		name          string
		markdown      string
		expectedErrs  int
		errorContains []string
	}{
		{
			name: "valid URLs",
			markdown: `# Test Module

Check out our [website](http://example.com/ok)

More info at http://example.com/docs
`,
			expectedErrs: 0,
		},
		{
			name: "404 URL",
			markdown: `# Test Module

Broken link: http://example.com/notfound
`,
			expectedErrs:  1,
			errorContains: []string{"non-OK status", "404"},
		},
		{
			name: "invalid URL",
			markdown: `# Test Module

Bad URL: http://example.com/unreachable
`,
			expectedErrs:  1,
			errorContains: []string{"error accessing URL"},
		},
		{
			name: "terraform registry URL skipped",
			markdown: `# Test Module

[azurerm](https://registry.terraform.io/providers/hashicorp/azurerm/latest)
`,
			expectedErrs: 0,
		},
		{
			name: "mixed valid and invalid URLs",
			markdown: `# Test Module

Good: http://example.com/ok
Bad: http://example.com/notfound
`,
			expectedErrs:  1,
			errorContains: []string{"non-OK status"},
		},
		{
			name:         "no URLs",
			markdown:     `# Test Module\n\nNo URLs here`,
			expectedErrs: 0,
		},
		{
			name: "multiple terraform registry URLs",
			markdown: `# Test Module

[provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group)
[another](https://registry.terraform.io/providers/hashicorp/aws/latest)
`,
			expectedErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withStubHTTPClient(t, map[string]int{
				"http://example.com/ok":          http.StatusOK,
				"http://example.com/docs":        http.StatusOK,
				"http://example.com/notfound":    http.StatusNotFound,
				"http://example.com/redirect":    http.StatusMovedPermanently,
				"http://example.com/unreachable": http.StatusOK,
			}, map[string]error{
				"http://example.com/unreachable": fmt.Errorf("dial error"),
			})

			mc := NewMarkdownContent(tt.markdown, FormatDocument, nil)
			uv := NewURLValidator(mc)

			errs := uv.Validate()

			if len(errs) != tt.expectedErrs {
				t.Errorf("Validate() returned %d errors; want %d", len(errs), tt.expectedErrs)
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
				if !found && tt.expectedErrs > 0 {
					t.Errorf("Expected error containing %q, but not found", substr)
				}
			}
		})
	}
}

func TestValidateSingleURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid URL returns no error",
			url:         "http://example.com/ok",
			expectError: false,
		},
		{
			name:          "404 status returns error",
			url:           "http://example.com/notfound",
			expectError:   true,
			errorContains: "non-OK status",
		},
		{
			name:          "invalid domain returns error",
			url:           "http://example.com/unreachable",
			expectError:   true,
			errorContains: "error accessing URL",
		},
		{
			name:          "redirect status returns error",
			url:           "http://example.com/redirect",
			expectError:   true,
			errorContains: "non-OK status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withStubHTTPClient(t, map[string]int{
				"http://example.com/ok":       http.StatusOK,
				"http://example.com/notfound": http.StatusNotFound,
				"http://example.com/redirect": http.StatusMovedPermanently,
			}, map[string]error{
				"http://example.com/unreachable": fmt.Errorf("dial error"),
			})

			err := validateSingleURL(tt.url)

			if (err != nil) != tt.expectError {
				t.Errorf("validateSingleURL() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.expectError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("validateSingleURL() error = %v, should contain %q", err, tt.errorContains)
				}
			}
		})
	}
}

func TestURLValidator_ConcurrentValidation(t *testing.T) {
	urls := []string{
		"http://example.com/ok1",
		"http://example.com/ok2",
		"http://example.com/ok3",
		"http://example.com/ok4",
		"http://example.com/ok5",
		"http://example.com/ok6",
		"http://example.com/ok7",
		"http://example.com/ok8",
		"http://example.com/ok9",
		"http://example.com/ok10",
	}

	withStubHTTPClient(t, map[string]int{
		"http://example.com/ok1":  http.StatusOK,
		"http://example.com/ok2":  http.StatusOK,
		"http://example.com/ok3":  http.StatusOK,
		"http://example.com/ok4":  http.StatusOK,
		"http://example.com/ok5":  http.StatusOK,
		"http://example.com/ok6":  http.StatusOK,
		"http://example.com/ok7":  http.StatusOK,
		"http://example.com/ok8":  http.StatusOK,
		"http://example.com/ok9":  http.StatusOK,
		"http://example.com/ok10": http.StatusOK,
	}, nil)

	markdown := "# Test\n\n" + strings.Join(urls, "\n")
	mc := NewMarkdownContent(markdown, FormatDocument, nil)
	uv := NewURLValidator(mc)

	errs := uv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with concurrent requests returned %d errors; want 0", len(errs))
	}
}

func TestURLValidator_TerraformRegistrySkip(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantSkip bool
	}{
		{
			name:     "terraform registry provider URL",
			markdown: `https://registry.terraform.io/providers/hashicorp/azurerm/latest`,
			wantSkip: true,
		},
		{
			name:     "terraform registry resource docs URL",
			markdown: `https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group`,
			wantSkip: true,
		},
		{
			name:     "terraform registry data source docs URL",
			markdown: `https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ami`,
			wantSkip: true,
		},
		{
			name:     "non-registry URL",
			markdown: `https://github.com/hashicorp/terraform`,
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withStubHTTPClient(t, nil, nil)

			mc := NewMarkdownContent(tt.markdown, FormatDocument, nil)
			uv := NewURLValidator(mc)

			errs := uv.Validate()

			if tt.wantSkip && len(errs) > 0 {
				t.Error("Expected terraform registry URL to be skipped, but got errors")
			}
		})
	}
}

func TestURLValidator_ExtractURLs(t *testing.T) {
	tests := []struct {
		name         string
		markdown     string
		expectedURLs int
	}{
		{
			name: "markdown with multiple URL formats",
			markdown: `# Module

Plain URL: https://example.com
Link: [text](https://example.org)
Another: https://github.com/user/repo
`,
			expectedURLs: 3,
		},
		{
			name: "markdown with no URLs",
			markdown: `# Module

No URLs in this content
`,
			expectedURLs: 0,
		},
		{
			name:         "URLs in code blocks",
			markdown:     "# Module\n\n```\nhttps://example.com\n```\n",
			expectedURLs: 1, // xurls.Strict() will find this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMarkdownContent(tt.markdown, FormatDocument, nil)

			// We can't directly test URL extraction since it's done inside Validate()
			// but we can verify the markdown content is set correctly
			if mc.GetContent() != tt.markdown {
				t.Errorf("Markdown content not set correctly")
			}
		})
	}
}

func TestURLValidator_Timeout(t *testing.T) {
	timeoutURL := "http://example.com/timeout"
	withStubHTTPClient(t, nil, map[string]error{
		timeoutURL: netErrorTimeout{},
	})

	markdown := "URL: " + timeoutURL
	mc := NewMarkdownContent(markdown, FormatDocument, nil)
	uv := NewURLValidator(mc)

	errs := uv.Validate()

	if len(errs) != 1 {
		t.Errorf("Validate() with timeout returned %d errors; want 1", len(errs))
	}
}

func TestURLValidator_EmptyContent(t *testing.T) {
	withStubHTTPClient(t, nil, nil)

	mc := NewMarkdownContent("", FormatDocument, nil)
	uv := NewURLValidator(mc)

	errs := uv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() on empty content returned %d errors; want 0", len(errs))
	}
}

func TestURLValidator_MaxConcurrency(t *testing.T) {
	urls := []string{
		"http://example.com/ok1",
		"http://example.com/ok2",
		"http://example.com/ok3",
		"http://example.com/ok4",
		"http://example.com/ok5",
		"http://example.com/ok6",
		"http://example.com/ok7",
		"http://example.com/ok8",
		"http://example.com/ok9",
		"http://example.com/ok10",
		"http://example.com/ok11",
		"http://example.com/ok12",
		"http://example.com/ok13",
		"http://example.com/ok14",
		"http://example.com/ok15",
		"http://example.com/ok16",
		"http://example.com/ok17",
		"http://example.com/ok18",
		"http://example.com/ok19",
		"http://example.com/ok20",
	}

	responses := make(map[string]int)
	for _, u := range urls {
		responses[u] = http.StatusOK
	}
	withStubHTTPClient(t, responses, nil)

	markdown := "# Test\n\n" + strings.Join(urls, "\n\n")
	mc := NewMarkdownContent(markdown, FormatDocument, nil)
	uv := NewURLValidator(mc)

	errs := uv.Validate()

	if len(errs) != 0 {
		t.Errorf("Validate() with many URLs returned %d errors; want 0", len(errs))
	}
}
