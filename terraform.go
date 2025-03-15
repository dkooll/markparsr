// package markparsr
//
// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"sync"
//
// 	"github.com/hashicorp/hcl/v2"
// 	"github.com/hashicorp/hcl/v2/hclparse"
// )
//
// // TerraformContent extracts resources, data sources, variables, and outputs
// // from Terraform files for documentation validation.
// type TerraformContent struct {
// 	workspace  string
// 	parserPool *sync.Pool
// 	fileCache  sync.Map
// }
//
// // NewTerraformContent creates a new analyzer for Terraform content.
// // Uses GITHUB_WORKSPACE environment variable as the root directory if available,
// // otherwise defaults to the current working directory.
// func NewTerraformContent() (*TerraformContent, error) {
// 	workspace := os.Getenv("GITHUB_WORKSPACE")
// 	if workspace == "" {
// 		var err error
// 		workspace, err = os.Getwd()
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to get current working directory: %w", err)
// 		}
// 	}
//
// 	return &TerraformContent{
// 		workspace: workspace,
// 		parserPool: &sync.Pool{
// 			New: func() any {
// 				return hclparse.NewParser()
// 			},
// 		},
// 	}, nil
// }
//
// // ExtractItems gets items of a specific block type (like variables or outputs)
// // from a Terraform file.
// func (tc *TerraformContent) ExtractItems(filePath, blockType string) ([]string, error) {
// 	content, err := os.ReadFile(filePath)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return []string{}, nil
// 		}
// 		return nil, fmt.Errorf("error reading file %s: %w", filepath.Base(filePath), err)
// 	}
//
// 	parser := tc.parserPool.Get().(*hclparse.Parser)
// 	defer tc.parserPool.Put(parser)
//
// 	file, parseDiags := parser.ParseHCL(content, filePath)
// 	if parseDiags.HasErrors() {
// 		return nil, fmt.Errorf("error parsing HCL in %s: %v", filepath.Base(filePath), parseDiags)
// 	}
//
// 	var items []string
// 	body := file.Body
// 	hclContent, _, diags := body.PartialContent(&hcl.BodySchema{
// 		Blocks: []hcl.BlockHeaderSchema{
// 			{Type: blockType, LabelNames: []string{"name"}},
// 		},
// 	})
//
// 	if diags.HasErrors() {
// 		return nil, fmt.Errorf("error getting content from %s: %v", filepath.Base(filePath), diags)
// 	}
//
// 	if hclContent == nil {
// 		return items, nil
// 	}
//
// 	for _, block := range hclContent.Blocks {
// 		if len(block.Labels) > 0 {
// 			itemName := strings.TrimSpace(block.Labels[0])
// 			items = append(items, itemName)
// 		}
// 	}
//
// 	return items, nil
// }
//
// // ExtractResourcesAndDataSources finds all resources and data sources defined in
// // Terraform files, searching in caller/main.tf and recursively through caller/modules.
// // It runs the search concurrently for better performance.
// func (tc *TerraformContent) ExtractResourcesAndDataSources() ([]string, []string, error) {
// 	var (
// 		resources      = make([]string, 0, 32)
// 		dataSources    = make([]string, 0, 32)
// 		resourceChan   = make(chan []string, 2)
// 		dataSourceChan = make(chan []string, 2)
// 		errChan        = make(chan error, 2)
// 		wg             sync.WaitGroup
// 	)
//
// 	wg.Add(2)
//
// 	// Extract from main.tf
// 	go func() {
// 		defer wg.Done()
// 		mainPath := filepath.Join(tc.workspace, "caller", "main.tf")
// 		specificResources, specificDataSources, err := tc.extractFromFilePath(mainPath)
// 		if err != nil && !os.IsNotExist(err) {
// 			errChan <- err
// 			return
// 		}
// 		resourceChan <- specificResources
// 		dataSourceChan <- specificDataSources
// 	}()
//
// 	// Extract from modules directory recursively
// 	go func() {
// 		defer wg.Done()
// 		modulesPath := filepath.Join(tc.workspace, "caller", "modules")
// 		modulesResources, modulesDataSources, err := tc.extractRecursively(modulesPath)
// 		if err != nil {
// 			errChan <- err
// 			return
// 		}
// 		resourceChan <- modulesResources
// 		dataSourceChan <- modulesDataSources
// 	}()
//
// 	go func() {
// 		wg.Wait()
// 		close(resourceChan)
// 		close(dataSourceChan)
// 		close(errChan)
// 	}()
//
// 	for r := range resourceChan {
// 		resources = append(resources, r...)
// 	}
// 	for ds := range dataSourceChan {
// 		dataSources = append(dataSources, ds...)
// 	}
//
// 	for err := range errChan {
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 	}
//
// 	return resources, dataSources, nil
// }
//
// // extractFromFilePath gets resources and data sources from a single Terraform file.
// func (tc *TerraformContent) extractFromFilePath(filePath string) ([]string, []string, error) {
// 	content, err := os.ReadFile(filePath)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return []string{}, []string{}, nil
// 		}
// 		return nil, nil, fmt.Errorf("error reading file %s: %w", filepath.Base(filePath), err)
// 	}
//
// 	parser := tc.parserPool.Get().(*hclparse.Parser)
// 	defer tc.parserPool.Put(parser)
//
// 	file, parseDiags := parser.ParseHCL(content, filePath)
// 	if parseDiags.HasErrors() {
// 		return nil, nil, fmt.Errorf("error parsing HCL in %s: %v", filepath.Base(filePath), parseDiags)
// 	}
//
// 	var resources []string
// 	var dataSources []string
// 	body := file.Body
// 	hclContent, _, diags := body.PartialContent(&hcl.BodySchema{
// 		Blocks: []hcl.BlockHeaderSchema{
// 			{Type: "resource", LabelNames: []string{"type", "name"}},
// 			{Type: "data", LabelNames: []string{"type", "name"}},
// 		},
// 	})
//
// 	if diags.HasErrors() {
// 		return nil, nil, fmt.Errorf("error getting content from %s: %v", filepath.Base(filePath), diags)
// 	}
//
// 	if hclContent == nil {
// 		return resources, dataSources, nil
// 	}
//
// 	for _, block := range hclContent.Blocks {
// 		if len(block.Labels) >= 2 {
// 			resourceType := strings.TrimSpace(block.Labels[0])
// 			resourceName := strings.TrimSpace(block.Labels[1])
// 			fullResourceName := resourceType + "." + resourceName
//
// 			switch block.Type {
// 			case "resource":
// 				resources = append(resources, resourceType)
// 				resources = append(resources, fullResourceName)
// 			case "data":
// 				dataSources = append(dataSources, resourceType)
// 				dataSources = append(dataSources, fullResourceName)
// 			}
// 		}
// 	}
//
// 	return resources, dataSources, nil
// }
//
// // extractRecursively walks through a directory and extracts resources
// // and data sources from all .tf files.
// func (tc *TerraformContent) extractRecursively(dirPath string) ([]string, []string, error) {
// 	var resources []string
// 	var dataSources []string
//
// 	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
// 		return resources, dataSources, nil
// 	} else if err != nil {
// 		return nil, nil, err
// 	}
//
// 	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if info.Mode().IsRegular() && filepath.Ext(path) == ".tf" {
// 			fileResources, fileDataSources, err := tc.extractFromFilePath(path)
// 			if err != nil {
// 				return err
// 			}
// 			resources = append(resources, fileResources...)
// 			dataSources = append(dataSources, fileDataSources...)
// 		}
// 		return nil
// 	})
//
// 	if err != nil {
// 		return nil, nil, err
// 	}
//
// 	return resources, dataSources, nil
// }

package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// TerraformContent extracts resources, data sources, variables, and outputs
// from Terraform files for documentation validation.
type TerraformContent struct {
	workspace  string
	parserPool *sync.Pool
	fileCache  sync.Map
}

// NewTerraformContent creates a new analyzer for Terraform content.
// It uses the provided module path as the root directory for Terraform files.
// For CI/CD compatibility, GITHUB_WORKSPACE is used only if modulePath is empty.
func NewTerraformContent(modulePath string) (*TerraformContent, error) {
	// If no modulePath provided, check for GITHUB_WORKSPACE
	if modulePath == "" {
		githubWorkspace := os.Getenv("GITHUB_WORKSPACE")
		if githubWorkspace != "" {
			modulePath = githubWorkspace
		} else {
			// Last resort - use current directory
			var err error
			modulePath, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current working directory: %w", err)
			}
		}
	}

	return &TerraformContent{
		workspace: modulePath,
		parserPool: &sync.Pool{
			New: func() any {
				return hclparse.NewParser()
			},
		},
	}, nil
}

// ExtractItems gets items of a specific block type from a Terraform file.
func (tc *TerraformContent) ExtractItems(filePath, blockType string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("error reading file %s: %w", filepath.Base(filePath), err)
	}

	parser := tc.parserPool.Get().(*hclparse.Parser)
	defer tc.parserPool.Put(parser)

	file, parseDiags := parser.ParseHCL(content, filePath)
	if parseDiags.HasErrors() {
		return nil, fmt.Errorf("error parsing HCL in %s: %v", filepath.Base(filePath), parseDiags)
	}

	var items []string
	body := file.Body
	hclContent, _, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: blockType, LabelNames: []string{"name"}},
		},
	})

	if diags.HasErrors() {
		return nil, fmt.Errorf("error getting content from %s: %v", filepath.Base(filePath), diags)
	}

	if hclContent == nil {
		return items, nil
	}

	for _, block := range hclContent.Blocks {
		if len(block.Labels) > 0 {
			itemName := strings.TrimSpace(block.Labels[0])
			items = append(items, itemName)
		}
	}

	return items, nil
}

// ExtractResourcesAndDataSources finds all resources and data sources defined in
// Terraform files, looking directly in the module directory.
func (tc *TerraformContent) ExtractResourcesAndDataSources() ([]string, []string, error) {
	var resources []string
	var dataSources []string

	// Scan all .tf files in the directory
	files, err := os.ReadDir(tc.workspace)
	if err != nil {
		if os.IsNotExist(err) {
			return resources, dataSources, nil
		}
		return nil, nil, fmt.Errorf("error reading directory %s: %w", tc.workspace, err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tf") {
			continue
		}

		filePath := filepath.Join(tc.workspace, file.Name())
		fileResources, fileDataSources, err := tc.extractFromFilePath(filePath)
		if err != nil {
			return nil, nil, err
		}

		resources = append(resources, fileResources...)
		dataSources = append(dataSources, fileDataSources...)
	}

	return resources, dataSources, nil
}

// extractFromFilePath gets resources and data sources from a single Terraform file.
func (tc *TerraformContent) extractFromFilePath(filePath string) ([]string, []string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, []string{}, nil
		}
		return nil, nil, fmt.Errorf("error reading file %s: %w", filepath.Base(filePath), err)
	}

	parser := tc.parserPool.Get().(*hclparse.Parser)
	defer tc.parserPool.Put(parser)

	file, parseDiags := parser.ParseHCL(content, filePath)
	if parseDiags.HasErrors() {
		return nil, nil, fmt.Errorf("error parsing HCL in %s: %v", filepath.Base(filePath), parseDiags)
	}

	var resources []string
	var dataSources []string
	body := file.Body
	hclContent, _, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}},
			{Type: "data", LabelNames: []string{"type", "name"}},
		},
	})

	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("error getting content from %s: %v", filepath.Base(filePath), diags)
	}

	if hclContent == nil {
		return resources, dataSources, nil
	}

	for _, block := range hclContent.Blocks {
		if len(block.Labels) >= 2 {
			resourceType := strings.TrimSpace(block.Labels[0])
			resourceName := strings.TrimSpace(block.Labels[1])
			fullResourceName := resourceType + "." + resourceName

			switch block.Type {
			case "resource":
				resources = append(resources, resourceType)
				resources = append(resources, fullResourceName)
			case "data":
				dataSources = append(dataSources, resourceType)
				dataSources = append(dataSources, fullResourceName)
			}
		}
	}

	return resources, dataSources, nil
}
