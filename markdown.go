package markparsr

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// MarkdownFormat represents the format of the Terraform documentation in markdown
type MarkdownFormat string

const (
	// FormatDocument represents terraform-docs document style output
	FormatDocument MarkdownFormat = "document"
	// FormatTable represents terraform-docs table style output
	FormatTable MarkdownFormat = "table"
	// FormatAuto will try to detect the format automatically
	FormatAuto MarkdownFormat = "auto"
)

// sectionConfig defines which columns are expected in a section's table
type sectionConfig struct {
	required []string
	optional []string
}

// FormatScore represents a score for a particular format detection
type FormatScore struct {
	Score       int
	SectionHits map[string]string // Maps section name to detected format
}

// MarkdownContent parses and analyzes Terraform module documentation
type MarkdownContent struct {
	data          string
	rootNode      ast.Node
	sections      map[string]bool
	sectionConfig map[string]sectionConfig
	format        MarkdownFormat
	stringPool    *sync.Pool
}

// NewMarkdownContent creates a new analyzer for markdown content
func NewMarkdownContent(data string, format MarkdownFormat) *MarkdownContent {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	rootNode := markdown.Parse([]byte(data), p)

	// Define section column requirements
	sectionConfig := map[string]sectionConfig{
		"Resources": {
			required: []string{"Name", "Type"},
		},
		"Providers": {
			required: []string{"Name", "Version"},
		},
		"Requirements": {
			required: []string{"Name", "Version"},
		},
		"Inputs": {
			required: []string{"Name", "Description", "Required"},
			optional: []string{"Type", "Default"},
		},
		"Required Inputs": {
			required: []string{"Name", "Description", "Required"},
			optional: []string{"Type", "Default"},
		},
		"Optional Inputs": {
			required: []string{"Name", "Description", "Required"},
			optional: []string{"Type", "Default"},
		},
		"Outputs": {
			required: []string{"Name", "Description"},
		},
	}

	mc := &MarkdownContent{
		data:     data,
		rootNode: rootNode,
		sections: make(map[string]bool),
		stringPool: &sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
		format:        format,
		sectionConfig: sectionConfig,
	}

	// Auto-detect format if not specified
	if format == FormatAuto {
		documentScore, tableScore, detectedFormat := mc.detectFormatHeuristic()
		mc.format = detectedFormat

		// Calculate confidence as a percentage
		totalScore := documentScore + tableScore
		var confidence float64
		if totalScore > 0 {
			if detectedFormat == FormatDocument {
				confidence = float64(documentScore) / float64(totalScore) * 100
			} else {
				confidence = float64(tableScore) / float64(totalScore) * 100
			}
		} else {
			confidence = 50.0 // No clear indicators, 50% confidence
		}

		// Always print the detection result with confidence
		fmt.Printf("Auto-detected markdown format: %s (confidence: %.1f%%)\n", mc.format, confidence)
	}

	return mc
}

// detectFormat determines whether the markdown uses document or table style
// using the heuristic approach
// detectFormatHeuristic determines markdown format using a scoring system
// Returns document score, table score, and the detected format
func (mc *MarkdownContent) detectFormatHeuristic() (int, int, MarkdownFormat) {
	// Create scoring structures
	documentScore := &FormatScore{
		Score:       0,
		SectionHits: make(map[string]string),
	}

	tableScore := &FormatScore{
		Score:       0,
		SectionHits: make(map[string]string),
	}

	// Key sections to analyze
	sectionsToCheck := []string{
		"Inputs", "Required Inputs", "Optional Inputs",
		"Outputs", "Resources", "Requirements", "Providers",
	}

	// Check each section independently
	for _, sectionName := range sectionsToCheck {
		format := mc.analyzeSection(sectionName)
		switch format {
		case FormatDocument:
			documentScore.Score++
			documentScore.SectionHits[sectionName] = "document"
		case FormatTable:
			tableScore.Score++
			tableScore.SectionHits[sectionName] = "table"
		}
	}

	// Check for L3 headings (strong indicator of document format)
	hasL3Headings := mc.hasLevel3Headings()
	if hasL3Headings {
		documentScore.Score += 2 // Stronger weight for L3 headings
	}

	// If there's a tie or no clear signal, apply secondary heuristics
	if documentScore.Score == tableScore.Score {
		// Check for code blocks with terraform types (more common in document format)
		if mc.hasTypeCodeBlocks() {
			documentScore.Score++
		}

		// Check for table-like structures (more common in table format)
		if mc.hasMultipleTableStructures() {
			tableScore.Score++
		}
	}

	// Determine winner
	var detectedFormat MarkdownFormat
	if documentScore.Score > tableScore.Score {
		detectedFormat = FormatDocument
	} else if tableScore.Score > documentScore.Score {
		detectedFormat = FormatTable
	} else {
		// If still tied, default to document format
		detectedFormat = FormatDocument
	}

	return documentScore.Score, tableScore.Score, detectedFormat
}

// analyzeSection checks a specific section for format indicators
func (mc *MarkdownContent) analyzeSection(sectionName string) MarkdownFormat {
	// First, try to find the section
	var sectionHeading *ast.Heading

	ast.WalkFunc(mc.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if heading, ok := node.(*ast.Heading); ok && heading.Level == 2 {
			headingText := strings.TrimSpace(mc.extractText(heading))
			if strings.EqualFold(headingText, sectionName) {
				sectionHeading = heading
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})

	if sectionHeading == nil {
		// Section not found
		return ""
	}

	// Look at what follows the heading
	next := getNextSibling(sectionHeading)
	if next == nil {
		return ""
	}

	// Check for table (strong indicator of table format)
	if _, isTable := next.(*ast.Table); isTable {
		return FormatTable
	}

	// Check for paragraph followed by L3 heading (indicator of document format)
	if _, isPara := next.(*ast.Paragraph); isPara {
		// See if there's a L3 heading after this
		afterPara := getNextSibling(next)
		if afterPara != nil {
			if heading, ok := afterPara.(*ast.Heading); ok && heading.Level == 3 {
				return FormatDocument
			}
		}
	}

	// Check directly for L3 heading (strong indicator of document format)
	if heading, ok := next.(*ast.Heading); ok && heading.Level == 3 {
		return FormatDocument
	}

	// If no clear indicators, return empty string
	return ""
}

// hasLevel3Headings checks if there are any level 3 headings in the document
func (mc *MarkdownContent) hasLevel3Headings() bool {
	hasL3 := false

	ast.WalkFunc(mc.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if heading, ok := node.(*ast.Heading); ok && heading.Level == 3 {
			hasL3 = true
			return ast.SkipChildren
		}
		return ast.GoToNext
	})

	return hasL3
}

// hasTypeCodeBlocks checks for code blocks with terraform types
func (mc *MarkdownContent) hasTypeCodeBlocks() bool {
	hasTypeBlocks := false

	ast.WalkFunc(mc.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if codeBlock, ok := node.(*ast.CodeBlock); ok {
			codeText := string(codeBlock.Literal)
			if strings.Contains(codeText, "object(") ||
				strings.Contains(codeText, "list(") ||
				strings.Contains(codeText, "map(") {
				hasTypeBlocks = true
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})

	return hasTypeBlocks
}

// hasMultipleTableStructures checks if there are multiple tables in the document
func (mc *MarkdownContent) hasMultipleTableStructures() bool {
	tableCount := 0

	ast.WalkFunc(mc.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if _, ok := node.(*ast.Table); ok {
			tableCount++
		}
		return ast.GoToNext
	})

	return tableCount > 1
}

// GetContent returns the full markdown content
func (mc *MarkdownContent) GetContent() string {
	return mc.data
}

// HasSection checks if a section exists in the markdown
func (mc *MarkdownContent) HasSection(sectionName string) bool {
	if found, exists := mc.sections[sectionName]; exists {
		return found
	}

	found := false
	ast.WalkFunc(mc.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering && heading.Level == 2 {
			text := strings.TrimSpace(mc.extractText(heading))
			if strings.EqualFold(text, sectionName) ||
				strings.EqualFold(text, sectionName+"s") ||
				(sectionName == "Inputs" && (strings.EqualFold(text, "Required Inputs") || strings.EqualFold(text, "Optional Inputs"))) {
				found = true
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})

	mc.sections[sectionName] = found
	return found
}

// GetAllSections returns a list of all H2 section names in the markdown
func (mc *MarkdownContent) GetAllSections() []string {
	var sections []string

	ast.WalkFunc(mc.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering && heading.Level == 2 {
			sectionName := strings.TrimSpace(mc.extractText(heading))
			if sectionName != "" {
				sections = append(sections, sectionName)
			}
		}
		return ast.GoToNext
	})

	return sections
}

// ExtractSectionItems extracts item names from a section, handling both document and table styles
func (mc *MarkdownContent) ExtractSectionItems(sectionNames ...string) []string {
	if mc.format == FormatTable {
		return mc.extractTableSectionItems(sectionNames...)
	}
	return mc.extractDocumentSectionItems(sectionNames...)
}

// extractDocumentSectionItems extracts item names from level 3 headings within specified sections
func (mc *MarkdownContent) extractDocumentSectionItems(sectionNames ...string) []string {
	var items []string
	inTargetSection := false

	ast.WalkFunc(mc.rootNode, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if heading, ok := n.(*ast.Heading); ok {
			headingText := strings.TrimSpace(mc.extractText(heading))
			if heading.Level == 2 {
				inTargetSection = false
				for _, sectionName := range sectionNames {
					if strings.EqualFold(headingText, sectionName) {
						inTargetSection = true
						break
					}
				}
			} else if heading.Level == 3 && inTargetSection {
				inputName := strings.Trim(headingText, " []")
				inputName = strings.TrimPrefix(inputName, "<a name=\"input_")
				inputName = strings.TrimPrefix(inputName, "<a name=\"output_")
				inputName = strings.TrimSuffix(inputName, "</a>")
				inputName = strings.TrimSuffix(inputName, "\"></a>")
				items = append(items, inputName)
			}
		}
		return ast.GoToNext
	})

	return items
}

// extractTableSectionItems extracts item names from tables within specified sections
func (mc *MarkdownContent) extractTableSectionItems(sectionNames ...string) []string {
	var items []string

	for _, sectionName := range sectionNames {
		sectionTable := mc.findSectionTable(sectionName)
		if sectionTable != nil {
			tableItems := mc.extractItemsFromTable(sectionTable)
			items = append(items, tableItems...)
		}
	}

	return items
}

// findSectionTable finds the table under a specific section heading
func (mc *MarkdownContent) findSectionTable(sectionName string) *ast.Table {
	var sectionTable *ast.Table

	ast.WalkFunc(mc.rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if heading, ok := node.(*ast.Heading); ok && heading.Level == 2 {
			headingText := strings.TrimSpace(mc.extractText(heading))
			// Match section name case-insensitive, including plurals
			if strings.EqualFold(headingText, sectionName) ||
				strings.EqualFold(headingText, sectionName+"s") {
				// Check the next node to see if it's a table
				next := getNextSibling(heading)
				if table, ok := next.(*ast.Table); ok {
					sectionTable = table
					return ast.SkipChildren
				}
			}
		}
		return ast.GoToNext
	})

	return sectionTable
}

// extractItemsFromTable extracts items from the first column of a markdown table
func (mc *MarkdownContent) extractItemsFromTable(table *ast.Table) []string {
	var items []string

	// Skip if empty table
	if len(table.GetChildren()) == 0 {
		return items
	}

	// Find the table body
	var bodyNode *ast.TableBody
	for _, child := range table.GetChildren() {
		if body, ok := child.(*ast.TableBody); ok {
			bodyNode = body
			break
		}
	}

	if bodyNode == nil {
		return items
	}

	// Process each row in the table body
	for _, rowChild := range bodyNode.GetChildren() {
		if row, ok := rowChild.(*ast.TableRow); ok && len(row.GetChildren()) > 0 {
			// Get the first cell (name column)
			if cell, ok := row.GetChildren()[0].(*ast.TableCell); ok {
				itemName := strings.TrimSpace(mc.extractText(cell))
				// Remove backticks if present (common in terraform-docs output)
				itemName = strings.Trim(itemName, "`")
				if itemName != "" && itemName != "Name" { // Skip header row
					items = append(items, itemName)
				}
			}
		}
	}

	return items
}

// ExtractResourcesAndDataSources finds Terraform resources and data sources in the markdown
func (mc *MarkdownContent) ExtractResourcesAndDataSources() ([]string, []string, error) {
	if mc.format == FormatTable {
		return mc.extractTableResourcesAndDataSources()
	}
	return mc.extractDocumentResourcesAndDataSources()
}

// extractDocumentResourcesAndDataSources finds resources in document style markdown
func (mc *MarkdownContent) extractDocumentResourcesAndDataSources() ([]string, []string, error) {
	var resources []string
	var dataSources []string
	inResourceSection := false

	ast.WalkFunc(mc.rootNode, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if heading, ok := n.(*ast.Heading); ok {
			headingText := mc.extractText(heading)
			if strings.Contains(headingText, "Resources") {
				inResourceSection = true
			} else if heading.Level <= 2 {
				inResourceSection = false
			}
		}

		if inResourceSection {
			if link, ok := n.(*ast.Link); ok {
				linkText := mc.extractText(link)
				destination := string(link.Destination)

				// Look for provider resource patterns (azurerm_, aws_, etc.)
				if hasProviderPrefix(linkText) {
					resourceName := strings.Split(linkText, "]")[0]
					resourceName = strings.TrimPrefix(resourceName, "[")
					baseName := strings.Split(resourceName, ".")[0]

					if strings.Contains(destination, "/data-sources/") {
						addUnique(&dataSources, resourceName)
						addUnique(&dataSources, baseName)
					} else {
						addUnique(&resources, resourceName)
						addUnique(&resources, baseName)
					}
				}
			}
		}
		return ast.GoToNext
	})

	if len(resources) == 0 && len(dataSources) == 0 {
		return nil, nil, fmt.Errorf("resources section not found or empty")
	}

	return resources, dataSources, nil
}

// extractTableResourcesAndDataSources finds resources in table style markdown
func (mc *MarkdownContent) extractTableResourcesAndDataSources() ([]string, []string, error) {
	var resources []string
	var dataSources []string

	// Find the Resources section table
	sectionTable := mc.findSectionTable("Resources")

	if sectionTable == nil {
		return nil, nil, fmt.Errorf("resources section not found or has no table")
	}

	// Find the Name and Type columns
	nameColIndex, typeColIndex := mc.findResourceTableColumns(sectionTable)
	if nameColIndex < 0 || typeColIndex < 0 {
		return nil, nil, fmt.Errorf("resources table is missing Name or Type columns")
	}

	// Extract resources from the table
	extractedResources, extractedDataSources := mc.extractResourcesFromTable(sectionTable, nameColIndex, typeColIndex)

	resources = append(resources, extractedResources...)
	dataSources = append(dataSources, extractedDataSources...)

	if len(resources) == 0 && len(dataSources) == 0 {
		return nil, nil, fmt.Errorf("resources section found but no resources or data sources extracted")
	}

	return resources, dataSources, nil
}

// findResourceTableColumns finds the column indices for Name and Type in a Resources table
func (mc *MarkdownContent) findResourceTableColumns(table *ast.Table) (nameIndex, typeIndex int) {
	nameIndex, typeIndex = -1, -1

	headers, err := mc.extractTableHeaders(table)
	if err != nil {
		return
	}

	// Find the Name and Type columns (case-insensitive)
	for i, header := range headers {
		if strings.EqualFold(header, "Name") {
			nameIndex = i
		} else if strings.EqualFold(header, "Type") {
			typeIndex = i
		}
	}

	return
}

// extractResourcesFromTable extracts resources and data sources from a table
func (mc *MarkdownContent) extractResourcesFromTable(table *ast.Table, nameColIndex, typeColIndex int) ([]string, []string) {
	var resources []string
	var dataSources []string

	// Find the table body
	var bodyNode *ast.TableBody
	for _, child := range table.GetChildren() {
		if body, ok := child.(*ast.TableBody); ok {
			bodyNode = body
			break
		}
	}

	if bodyNode == nil {
		return resources, dataSources
	}

	// Process each row in the table
	for _, rowNode := range bodyNode.GetChildren() {
		if row, ok := rowNode.(*ast.TableRow); ok {
			cells := row.GetChildren()
			if len(cells) <= max(nameColIndex, typeColIndex) {
				continue
			}

			// Get the resource name and type
			nameCell, nameOk := cells[nameColIndex].(*ast.TableCell)
			typeCell, typeOk := cells[typeColIndex].(*ast.TableCell)

			if !nameOk || !typeOk {
				continue
			}

			name := strings.TrimSpace(mc.extractText(nameCell))
			resourceType := strings.ToLower(strings.TrimSpace(mc.extractText(typeCell)))

			// Clean up the name (remove backticks, brackets, etc.)
			name = strings.Trim(name, "` []")
			if name == "" || name == "Name" {
				continue // Skip header row or empty rows
			}

			// Extract the base resource type
			parts := strings.Split(name, ".")
			baseName := parts[0]

			if hasProviderPrefix(name) {
				if strings.Contains(resourceType, "data source") {
					addUnique(&dataSources, name)
					addUnique(&dataSources, baseName)
				} else {
					addUnique(&resources, name)
					addUnique(&resources, baseName)
				}
			}
		}
	}

	return resources, dataSources
}

// ValidateTableColumns checks that all tables in sections have the required columns
func (mc *MarkdownContent) ValidateTableColumns() []error {
	var allErrors []error

	// Only validate if we're in table format
	if mc.format != FormatTable {
		return allErrors
	}

	// Validate each section that has column requirements
	for sectionName, config := range mc.sectionConfig {
		if len(config.required) == 0 && len(config.optional) == 0 {
			continue
		}

		// Find the section's table
		sectionTable := mc.findSectionTable(sectionName)
		if sectionTable == nil {
			continue
		}

		// Extract and validate column headers
		headers, err := mc.extractTableHeaders(sectionTable)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("%s: %v", sectionName, err))
			continue
		}

		// Validate columns
		errors := mc.validateColumns(sectionName, config.required, config.optional, headers)
		allErrors = append(allErrors, errors...)
	}

	return allErrors
}

// extractTableHeaders gets the header row of a table
func (mc *MarkdownContent) extractTableHeaders(table *ast.Table) ([]string, error) {
	var headers []string

	if len(table.GetChildren()) == 0 {
		return nil, fmt.Errorf("table is empty")
	}

	// Find the header node
	var headerNode *ast.TableHeader
	for _, child := range table.GetChildren() {
		if h, ok := child.(*ast.TableHeader); ok {
			headerNode = h
			break
		}
	}

	if headerNode == nil {
		return nil, fmt.Errorf("table has no header row")
	}

	// Get headers from the header row
	for _, rowNode := range headerNode.GetChildren() {
		if row, ok := rowNode.(*ast.TableRow); ok {
			for _, cellNode := range row.GetChildren() {
				if cell, ok := cellNode.(*ast.TableCell); ok {
					headerText := strings.TrimSpace(mc.extractText(cell))
					headers = append(headers, headerText)
				}
			}
		}
	}

	return headers, nil
}

// validateColumns checks if required columns are present and alerts about typos
func (mc *MarkdownContent) validateColumns(sectionName string, required, optional, actual []string) []error {
	var errors []error

	// Create maps for valid columns
	requiredLower := make(map[string]string) // lowercase → original
	optionalLower := make(map[string]string) // lowercase → original

	for _, col := range required {
		requiredLower[strings.ToLower(col)] = col
	}
	for _, col := range optional {
		optionalLower[strings.ToLower(col)] = col
	}

	// Track which required columns were found, including close matches
	foundRequiredExact := make(map[string]bool) // Exact matches by lowercase name
	foundRequiredClose := make(map[string]bool) // Close matches (typos) by lowercase name

	// First pass: identify typos and exact matches
	for _, actual := range actual {
		actualLower := strings.ToLower(actual)

		// Check if this is a required column (exact match)
		if originalRequired, isRequired := requiredLower[actualLower]; isRequired {
			foundRequiredExact[actualLower] = true

			// Check for capitalization issues
			if actual != originalRequired {
				errors = append(errors, fmt.Errorf(
					"%s: column '%s' should be '%s'",
					sectionName, actual, originalRequired))
			}
			continue
		}

		// Check if this is an optional column (exact match)
		if originalOptional, isOptional := optionalLower[actualLower]; isOptional {
			// Check for capitalization issues
			if actual != originalOptional {
				errors = append(errors, fmt.Errorf(
					"%s: column '%s' should be '%s'",
					sectionName, actual, originalOptional))
			}
			continue
		}

		// Look for close matches to required columns
		isCloseMatch := false
		for reqLower, reqOriginal := range requiredLower {
			if isSimilarColumn(actualLower, reqLower) {
				errors = append(errors, fmt.Errorf(
					"%s: unexpected column '%s' (did you mean '%s'?)",
					sectionName, actual, reqOriginal))
				foundRequiredClose[reqLower] = true
				isCloseMatch = true
				break
			}
		}

		if isCloseMatch {
			continue
		}

		// Look for close matches to optional columns
		isCloseMatch = false
		for optLower, optOriginal := range optionalLower {
			if isSimilarColumn(actualLower, optLower) {
				errors = append(errors, fmt.Errorf(
					"%s: unexpected column '%s' (did you mean '%s'?)",
					sectionName, actual, optOriginal))
				isCloseMatch = true
				break
			}
		}

		if isCloseMatch {
			continue
		}

		// Not a match for anything
		errors = append(errors, fmt.Errorf(
			"%s: unexpected column '%s'", sectionName, actual))
	}

	// Check for missing required columns - only if not found as either exact or close match
	for reqLower, reqOriginal := range requiredLower {
		if !foundRequiredExact[reqLower] && !foundRequiredClose[reqLower] {
			errors = append(errors, fmt.Errorf(
				"%s: missing required column '%s'", sectionName, reqOriginal))
		}
	}

	return errors
}

// isSimilarColumn checks if two column names are similar (likely typos)
func isSimilarColumn(a, b string) bool {
	// Common variations
	if a+"s" == b || a == b+"s" {
		return true
	}

	// Edit distance for typos
	if levenshtein(a, b) <= 2 {
		return true
	}

	return false
}

// extractText gets the text content from a node, using a string pool for efficiency
func (mc *MarkdownContent) extractText(node ast.Node) string {
	sb := mc.stringPool.Get().(*strings.Builder)
	sb.Reset()
	defer mc.stringPool.Put(sb)

	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch tn := n.(type) {
			case *ast.Text:
				sb.Write(tn.Literal)
			case *ast.Code:
				sb.Write(tn.Literal)
			}
		}
		return ast.GoToNext
	})

	return sb.String()
}

// Helper functions

// getNextSibling returns the next sibling of a node
func getNextSibling(node ast.Node) ast.Node {
	parent := node.GetParent()
	if parent == nil {
		return nil
	}
	children := parent.GetChildren()
	for i, n := range children {
		if n == node && i+1 < len(children) {
			return children[i+1]
		}
	}
	return nil
}

// hasProviderPrefix checks if a string has a recognized provider prefix
func hasProviderPrefix(s string) bool {
	s = strings.ToLower(s)
	commonPrefixes := []string{
		"azurerm_", "random_",
	}
	for _, prefix := range commonPrefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// addUnique adds a string to a slice if it's not already present
func addUnique(slice *[]string, item string) {
	if !slices.Contains(*slice, item) {
		*slice = append(*slice, item)
	}
}

// levenshtein calculates the edit distance between two strings
func levenshtein(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create two work vectors of integer distances
	v0 := make([]int, len(s2)+1)
	v1 := make([]int, len(s2)+1)

	// Initialize v0 (previous row of distances)
	for i := range v0 {
		v0[i] = i
	}

	// Calculate rows
	for i := range s1 {
		// First element of v1 is A[i+1][0]
		v1[0] = i + 1

		// Calculate column entries
		for j := range s2 {
			cost := 1
			if s1[i] == s2[j] {
				cost = 0
			}
			v1[j+1] = min(v1[j]+1, v0[j+1]+1, v0[j]+cost)
		}

		// Copy v1 to v0 for next iteration
		copy(v0, v1)
	}

	return v1[len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
