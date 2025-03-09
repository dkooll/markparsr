package markparsr

import (
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// formatError formats an error message
func formatError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// compareHeaders compares expected and actual headers
func compareHeaders(expected, actual string) error {
	if expected != actual {
		if actual == "" {
			return formatError("incorrect header:\n  expected '%s', found 'not present'", expected)
		}
		return formatError("incorrect header:\n  expected '%s', found '%s'", expected, actual)
	}
	return nil
}

// findMissingItems finds items that are in a but not in b
func findMissingItems(a, b []string) []string {
	bSet := make(map[string]struct{}, len(b))
	for _, x := range b {
		bSet[x] = struct{}{}
	}
	var missing []string
	for _, x := range a {
		if _, found := bSet[x]; !found {
			missing = append(missing, x)
		}
	}
	return missing
}

// validateColumns checks if the required and optional columns are present
func validateColumns(header string, required, optional, actual []string) []error {
	var errors []error

	// Create a map of valid columns
	validColumns := make(map[string]bool)
	for _, col := range required {
		validColumns[col] = true
	}
	for _, col := range optional {
		validColumns[col] = true
	}

	// Track found and invalid columns
	foundColumns := make(map[string]bool)
	hasInvalidColumns := false

	// First check for unexpected columns
	for _, act := range actual {
		if !validColumns[act] {
			hasInvalidColumns = true
			errors = append(errors, formatError("unexpected column '%s' in table under header: %s", act, header))
		}
		foundColumns[act] = true
	}

	// Only check for missing required columns if there were no invalid columns
	if !hasInvalidColumns {
		for _, req := range required {
			if !foundColumns[req] {
				errors = append(errors, formatError("missing required column '%s' in table under header: %s", req, header))
			}
		}
	}

	return errors
}

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

// extractTableHeaders extracts headers from a markdown table
func extractTableHeaders(table *ast.Table) ([]string, error) {
	headers := []string{}

	if len(table.GetChildren()) == 0 {
		return nil, fmt.Errorf("table is empty")
	}

	// The first child should be TableHeader
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

	// The header row is under TableHeader
	for _, rowNode := range headerNode.GetChildren() {
		if row, ok := rowNode.(*ast.TableRow); ok {
			for _, cellNode := range row.GetChildren() {
				if cell, ok := cellNode.(*ast.TableCell); ok {
					headerText := strings.TrimSpace(extractTextFromNodes(cell.GetChildren()))
					headers = append(headers, headerText)
				}
			}
		}
	}

	return headers, nil
}

// extractText extracts text from a node, including code spans
func extractText(node ast.Node) string {
	var sb strings.Builder
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

// extractTextFromNodes extracts text from a slice of nodes
func extractTextFromNodes(nodes []ast.Node) string {
	var sb strings.Builder
	for _, node := range nodes {
		sb.WriteString(extractText(node))
	}
	return sb.String()
}
