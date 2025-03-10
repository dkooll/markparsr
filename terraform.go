package markparsr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// TerraformDefinitionValidator validates Terraform definitions
type TerraformDefinitionValidator struct {
	data string
}

// NewTerraformDefinitionValidator creates a new TerraformDefinitionValidator
func NewTerraformDefinitionValidator(data string) *TerraformDefinitionValidator {
	return &TerraformDefinitionValidator{data: data}
}

// Validate compares Terraform resources with those documented in the markdown
func (tdv *TerraformDefinitionValidator) Validate() []error {
	tfResources, tfDataSources, err := extractTerraformResources()
	if err != nil {
		return []error{err}
	}

	readmeResources, readmeDataSources, err := extractReadmeResources(tdv.data)
	if err != nil {
		return []error{err}
	}

	var errors []error
	errors = append(errors, compareTerraformAndMarkdown(tfResources, readmeResources, "Resources")...)
	errors = append(errors, compareTerraformAndMarkdown(tfDataSources, readmeDataSources, "Data Sources")...)

	return errors
}

// extractReadmeResources extracts resources and data sources from the markdown
func extractReadmeResources(data string) ([]string, []string, error) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	rootNode := markdown.Parse([]byte(data), p)

	var resources []string
	var dataSources []string
	var inResourcesSection bool

	ast.WalkFunc(rootNode, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering && heading.Level == 2 {
			text := strings.TrimSpace(extractText(heading))
			if strings.EqualFold(text, "Resources") {
				inResourcesSection = true
				return ast.GoToNext
			}
			inResourcesSection = false
		}

		if inResourcesSection {
			if table, ok := node.(*ast.Table); ok && entering {
				// Extract items from the table
				var bodyNode *ast.TableBody
				for _, child := range table.GetChildren() {
					if body, ok := child.(*ast.TableBody); ok {
						bodyNode = body
						break
					}
				}
				if bodyNode == nil {
					return ast.GoToNext
				}

				for _, rowChild := range bodyNode.GetChildren() {
					if tableRow, ok := rowChild.(*ast.TableRow); ok {
						cells := tableRow.GetChildren()
						if len(cells) >= 2 {
							nameCell, ok1 := cells[0].(*ast.TableCell)
							typeCell, ok2 := cells[1].(*ast.TableCell)
							if ok1 && ok2 {
								name := extractTextFromNodes(nameCell.GetChildren())
								name = strings.TrimSpace(name)
								name = strings.Trim(name, "[]") // Remove brackets
								name = strings.TrimSpace(name)
								resourceType := extractTextFromNodes(typeCell.GetChildren())
								resourceType = strings.TrimSpace(resourceType)
								if strings.EqualFold(resourceType, "resource") {
									resources = append(resources, name)
								} else if strings.EqualFold(resourceType, "data source") {
									dataSources = append(dataSources, name)
								}
							}
						}
					}
				}
				inResourcesSection = false // We've processed the table, exit the section
				return ast.SkipChildren
			}
		}
		return ast.GoToNext
	})

	if len(resources) == 0 && len(dataSources) == 0 {
		return nil, nil, errors.New("resources section not found or empty")
	}

	return resources, dataSources, nil
}

// extractTerraformResources extracts resources and data sources from Terraform files
func extractTerraformResources() ([]string, []string, error) {
	var resources []string
	var dataSources []string

	workspace := os.Getenv("GITHUB_WORKSPACE")
	if workspace == "" {
		var err error
		workspace, err = os.Getwd()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get current working directory: %v", err)
		}
	}

	callerPath := filepath.Join(workspace, "caller")
	allResources, allDataSources, err := extractRecursively(callerPath)
	if err != nil {
		return nil, nil, err
	}

	resources = append(resources, allResources...)
	dataSources = append(dataSources, allDataSources...)

	return resources, dataSources, nil
}

// extractRecursively extracts resources and data sources recursively, skipping specified directories
func extractRecursively(dirPath string) ([]string, []string, error) {
	var resources []string
	var dataSources []string
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return resources, dataSources, nil
	} else if err != nil {
		return nil, nil, err
	}

	// Directories to skip
	skipDirs := map[string]struct{}{
		"modules":  {},
		"examples": {},
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the modules and examples directories
		if info.IsDir() {
			if _, shouldSkip := skipDirs[info.Name()]; shouldSkip {
				return filepath.SkipDir
			}
		}

		if info.Mode().IsRegular() && filepath.Ext(path) == ".tf" {
			fileResources, fileDataSources, err := extractFromFilePath(path)
			if err != nil {
				return err
			}
			resources = append(resources, fileResources...)
			dataSources = append(dataSources, fileDataSources...)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return resources, dataSources, nil
}

// extractFromFilePath extracts resources and data sources from a Terraform file
func extractFromFilePath(filePath string) ([]string, []string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file %s: %v", filepath.Base(filePath), err)
	}

	parser := hclparse.NewParser()
	file, parseDiags := parser.ParseHCL(content, filePath)
	if parseDiags.HasErrors() {
		return nil, nil, fmt.Errorf("error parsing HCL in %s: %v", filepath.Base(filePath), parseDiags)
	}

	var resources []string
	var dataSources []string
	body := file.Body

	// Initialize diagnostics variable
	var diags hcl.Diagnostics

	// Use PartialContent to allow unknown blocks
	hclContent, _, contentDiags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}},
			{Type: "data", LabelNames: []string{"type", "name"}},
		},
	})

	// Append diagnostics
	diags = append(diags, contentDiags...)

	// Filter out diagnostics related to unsupported block types
	diags = filterUnsupportedBlockDiagnostics(diags)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("error getting content from %s: %v", filepath.Base(filePath), diags)
	}

	if hclContent == nil {
		// No relevant blocks found
		return resources, dataSources, nil
	}

	for _, block := range hclContent.Blocks {
		if len(block.Labels) >= 2 {
			resourceType := strings.TrimSpace(block.Labels[0])
			resourceName := strings.TrimSpace(block.Labels[1])
			fullResourceName := resourceType + "." + resourceName

			switch block.Type {
			case "resource":
				resources = append(resources, fullResourceName)
			case "data":
				dataSources = append(dataSources, fullResourceName)
			}
		}
	}

	return resources, dataSources, nil
}
