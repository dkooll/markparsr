package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type defaultFileReader struct{}

func (dfr *defaultFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

type defaultHCLParser struct{}

func (dhp *defaultHCLParser) ParseHCL(content []byte, filename string) (*hcl.File, hcl.Diagnostics) {
	parser := hclparse.NewParser()
	return parser.ParseHCL(content, filename)
}

type TerraformContent struct {
	workspace  string
	fileReader FileReader
	hclParser  HCLParser
	readDir    func(string) ([]os.DirEntry, error)
}

func NewTerraformContent(modulePath string) (*TerraformContent, error) {
	if modulePath == "" {
		githubWorkspace := os.Getenv("GITHUB_WORKSPACE")
		if githubWorkspace != "" {
			modulePath = githubWorkspace
		} else {
			var err error
			modulePath, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current working directory: %w", err)
			}
		}
	}

	return &TerraformContent{
		workspace:  modulePath,
		fileReader: &defaultFileReader{},
		hclParser:  &defaultHCLParser{},
		readDir:    os.ReadDir,
	}, nil
}

func (tc *TerraformContent) parseFile(filePath string) (*hcl.File, error) {
	content, err := tc.fileReader.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading file %s: %w", filepath.Base(filePath), err)
	}

	file, parseDiags := tc.hclParser.ParseHCL(content, filePath)
	if parseDiags.HasErrors() {
		return nil, fmt.Errorf("error parsing HCL in %s: %v", filepath.Base(filePath), parseDiags)
	}

	return file, nil
}

func (tc *TerraformContent) ExtractItems(filePath, blockType string) ([]string, error) {
	file, err := tc.parseFile(filePath)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return []string{}, nil
	}

	return tc.extractItemsFromFile(file, filePath, blockType)
}

func (tc *TerraformContent) extractItemsFromFile(file *hcl.File, filePath, blockType string) ([]string, error) {
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

func (tc *TerraformContent) ExtractModuleItems(blockType string) ([]string, error) {
	files, err := tc.readDir(tc.workspace)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("error reading directory %s: %w", tc.workspace, err)
	}

	seen := make(map[string]struct{})
	var items []string

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tf") {
			continue
		}

		filePath := filepath.Join(tc.workspace, file.Name())
		fileItems, err := tc.ExtractItems(filePath, blockType)
		if err != nil {
			return nil, err
		}

		for _, item := range fileItems {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}

	return items, nil
}

func (tc *TerraformContent) ExtractResourcesAndDataSources() ([]string, []string, error) {
	var resources []string
	var dataSources []string

	files, err := tc.readDir(tc.workspace)
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

func (tc *TerraformContent) extractFromFilePath(filePath string) ([]string, []string, error) {
	file, err := tc.parseFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	if file == nil {
		return []string{}, []string{}, nil
	}

	return tc.extractResourcesFromFile(file, filePath)
}

func (tc *TerraformContent) extractResourcesFromFile(file *hcl.File, filePath string) ([]string, []string, error) {
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
