package markparsr

import (
	"fmt"
	"strings"
)

type defaultComparisonValidator struct{}

func NewComparisonValidator() ComparisonValidator {
	return &defaultComparisonValidator{}
}

func (dcv *defaultComparisonValidator) ValidateItems(tfItems, mdItems []string, itemType string) []error {
	return compareTerraformAndMarkdown(tfItems, mdItems, itemType)
}

type normalizedItem struct {
	original string
	key      string
	base     string
	hasDot   bool
}

type itemIndex struct {
	entries map[string]normalizedItem
	byBase  map[string][]normalizedItem
}

func buildItemIndex(items []string) itemIndex {
	index := itemIndex{
		entries: make(map[string]normalizedItem),
		byBase:  make(map[string][]normalizedItem),
	}
	qualifiedBases := make(map[string]bool)

	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, exists := index.entries[key]; exists {
			continue
		}

		base := key
		hasDot := false
		if pos := strings.Index(trimmed, "."); pos != -1 {
			hasDot = true
			base = strings.ToLower(strings.TrimSpace(trimmed[:pos]))
			qualifiedBases[base] = true
		}

		entry := normalizedItem{
			original: trimmed,
			key:      key,
			base:     base,
			hasDot:   hasDot,
		}

		index.entries[key] = entry
		index.byBase[base] = append(index.byBase[base], entry)
	}

	for key, entry := range index.entries {
		if !entry.hasDot && qualifiedBases[entry.base] {
			delete(index.entries, key)
		}
	}

	index.byBase = make(map[string][]normalizedItem)
	for _, entry := range index.entries {
		index.byBase[entry.base] = append(index.byBase[entry.base], entry)
	}

	return index
}

func (idx itemIndex) items() []normalizedItem {
	result := make([]normalizedItem, 0, len(idx.entries))
	for _, entry := range idx.entries {
		result = append(result, entry)
	}
	return result
}

func (idx itemIndex) hasMatch(target normalizedItem) bool {
	if _, ok := idx.entries[target.key]; ok {
		return true
	}
	if entries, ok := idx.byBase[target.base]; ok && len(entries) > 0 {
		return true
	}
	return false
}

func compareTerraformAndMarkdown(tfItems, mdItems []string, itemType string) []error {
	tfIndex := buildItemIndex(tfItems)
	mdIndex := buildItemIndex(mdItems)

	var errors []error

	for _, entry := range tfIndex.items() {
		if mdIndex.hasMatch(entry) {
			continue
		}
		errors = append(errors, fmt.Errorf("%s in Terraform but missing in markdown: %s", itemType, entry.original))
	}

	for _, entry := range mdIndex.items() {
		if tfIndex.hasMatch(entry) {
			continue
		}
		errors = append(errors, fmt.Errorf("%s in markdown but missing in Terraform: %s", itemType, entry.original))
	}

	return errors
}
