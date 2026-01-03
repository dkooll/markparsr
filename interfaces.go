package markparsr

import "github.com/hashicorp/hcl/v2"

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type HCLParser interface {
	ParseHCL(content []byte, filename string) (*hcl.File, hcl.Diagnostics)
}

type ResourceExtractor interface {
	ExtractResourcesAndDataSources() ([]string, []string, error)
	ExtractItems(filePath, blockType string) ([]string, error)
}

type DocumentParser interface {
	GetContent() string
	GetAllSections() []string
	HasSection(sectionName string) bool
}

type SectionExtractor interface {
	ExtractSectionItems(sectionNames ...string) []string
}

type ResourceDocumentExtractor interface {
	ExtractResourcesAndDataSources() ([]string, []string, error)
}

type ComparisonValidator interface {
	ValidateItems(tfItems, mdItems []string, itemType string) []error
}

type StringUtils interface {
	LevenshteinDistance(s1, s2 string) int
	IsSimilarSection(found, expected string) bool
}

type Validator interface {
	Validate() []error
}
