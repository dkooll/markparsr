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

type TerraformContent struct {
	workspace  string
	parserPool *sync.Pool
	fileCache  sync.Map
}

func NewTerraformContent() (*TerraformContent, error) {
	workspace := os.Getenv("GITHUB_WORKSPACE")
	if workspace == "" {
		var err error
		workspace, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	return &TerraformContent{
		workspace: workspace,
		parserPool: &sync.Pool{
			New: func() any {
				return hclparse.NewParser()
			},
		},
	}, nil
}

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

func (tc *TerraformContent) ExtractResourcesAndDataSources() ([]string, []string, error) {
	var (
		resources      = make([]string, 0, 32)
		dataSources    = make([]string, 0, 32)
		resourceChan   = make(chan []string, 2)
		dataSourceChan = make(chan []string, 2)
		errChan        = make(chan error, 2)
		wg             sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		mainPath := filepath.Join(tc.workspace, "caller", "main.tf")
		specificResources, specificDataSources, err := tc.extractFromFilePath(mainPath)
		if err != nil && !os.IsNotExist(err) {
			errChan <- err
			return
		}
		resourceChan <- specificResources
		dataSourceChan <- specificDataSources
	}()

	go func() {
		defer wg.Done()
		modulesPath := filepath.Join(tc.workspace, "caller", "modules")
		modulesResources, modulesDataSources, err := tc.extractRecursively(modulesPath)
		if err != nil {
			errChan <- err
			return
		}
		resourceChan <- modulesResources
		dataSourceChan <- modulesDataSources
	}()

	go func() {
		wg.Wait()
		close(resourceChan)
		close(dataSourceChan)
		close(errChan)
	}()

	for r := range resourceChan {
		resources = append(resources, r...)
	}
	for ds := range dataSourceChan {
		dataSources = append(dataSources, ds...)
	}

	for err := range errChan {
		if err != nil {
			return nil, nil, err
		}
	}

	return resources, dataSources, nil
}

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

func (tc *TerraformContent) extractRecursively(dirPath string) ([]string, []string, error) {
	var resources []string
	var dataSources []string

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return resources, dataSources, nil
	} else if err != nil {
		return nil, nil, err
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && filepath.Ext(path) == ".tf" {
			fileResources, fileDataSources, err := tc.extractFromFilePath(path)
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
