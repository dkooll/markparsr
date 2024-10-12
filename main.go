package main

import (
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// Function to extract text from AST nodes
func extractText(node ast.Node) string {
	var sb strings.Builder
	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if textNode, ok := n.(*ast.Text); ok {
			sb.Write(textNode.Literal)
		}
		return ast.GoToNext
	})
	return sb.String()
}

// Function to print all headings in the markdown content
func printHeadings(node ast.Node) {
	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := n.(*ast.Heading); ok && entering {
			fmt.Printf("Heading Level %d: %s\n", heading.Level, extractText(heading))
		}
		return ast.GoToNext
	})
}

// Function to print the AST tree structure with dynamic type info and level
func printASTTree(node ast.Node, depth int) {
	indent := strings.Repeat("  ", depth) // Indentation based on depth
	fmt.Printf("%sNode Type: %T (Level: %d)\n", indent, node, depth) // Print the dynamic type of the node and its level

	// Recursively walk through the children
	for _, child := range node.GetChildren() {
		printASTTree(child, depth+1) // Increase depth for children
	}
}

// Function to extract resource names from markdown list items
func extractResourceNames(node ast.Node) {
	fmt.Println("\nExtracted Resources:")
	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if listItem, ok := n.(*ast.ListItem); ok && entering {
			itemText := extractText(listItem)
			if strings.Contains(itemText, "azurerm_") {
				resource := strings.Split(itemText, " ")[0]
				fmt.Println(resource) // Print the resource name
			}
		}
		return ast.GoToNext
	})
}

func main() {
	// Sample markdown content with headings and resources
	markdownContent := []byte(`
# Title of Document
Some introductory text.

# Another Heading.

## Resources
The following resources are used by this module:

- [azurerm_kubernetes_cluster.this](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/kubernetes_cluster) (resource)
- [azurerm_kubernetes_cluster_extension.this](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/kubernetes_cluster_extension) (resource)
- [azurerm_kubernetes_cluster_node_pool.this](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/kubernetes_cluster_node_pool) (resource)
- [azurerm_role_assignment.this](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/role_assignment) (resource)
- [azurerm_user_assigned_identity.this](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/user_assigned_identity) (resource)
`)

	// Parse the markdown content into an AST
	mdParser := parser.New()
	node := markdown.Parse(markdownContent, mdParser)

	// Call function to print headings
	fmt.Println("Printing Headings:")
	printHeadings(node)

	// Call function to print the entire AST tree
	fmt.Println("\nPrinting AST Tree:")
	printASTTree(node, 0)

	// Call function to extract and print resource names
	extractResourceNames(node)
}
