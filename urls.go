package markparsr

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"mvdan.cc/xurls/v2"
)

// URLValidator checks that all URLs in markdown documentation are accessible.
type URLValidator struct {
	content *MarkdownContent
}

// NewURLValidator creates a validator for checking URLs in markdown content.
func NewURLValidator(content *MarkdownContent) *URLValidator {
	return &URLValidator{content: content}
}

// Validate checks all URLs in the markdown content to ensure they are accessible.
// Skips Terraform Registry URLs (registry.terraform.io/providers/) and makes
// concurrent HTTP requests to each URL for better performance.
func (uv *URLValidator) Validate() []error {
	rxStrict := xurls.Strict()
	urls := rxStrict.FindAllString(uv.content.data, -1)

	var wg sync.WaitGroup
	errChan := make(chan error, len(urls))

	for _, u := range urls {
		if strings.Contains(u, "registry.terraform.io/providers/") {
			continue
		}
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := validateSingleURL(url); err != nil {
				errChan <- err
			}
		}(u)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	return errors
}

// validateSingleURL checks if a URL is accessible and returns a 200 OK status.
func validateSingleURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error accessing URL: %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("URL returned non-OK status: %s: Status: %d", url, resp.StatusCode)
	}
	return nil
}
