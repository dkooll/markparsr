package markparsr

// TerraformDefinitionValidator ensures that all resources and data sources
// in Terraform code are documented in markdown, and vice versa.
type TerraformDefinitionValidator struct {
	markdown  *MarkdownContent
	terraform *TerraformContent
}

// NewTerraformDefinitionValidator creates a validator to compare resources and
// data sources between Terraform code and markdown documentation.
func NewTerraformDefinitionValidator(markdown *MarkdownContent, terraform *TerraformContent) *TerraformDefinitionValidator {
	return &TerraformDefinitionValidator{
		markdown:  markdown,
		terraform: terraform,
	}
}

// Validate compares resources and data sources between Terraform and markdown.
// It reports any resources that are in code but not documented, and vice versa.
func (tdv *TerraformDefinitionValidator) Validate() []error {
	tfResources, tfDataSources, err := tdv.terraform.ExtractResourcesAndDataSources()
	if err != nil {
		return []error{err}
	}

	readmeResources, readmeDataSources, err := tdv.markdown.ExtractResourcesAndDataSources()
	if err != nil {
		return []error{err}
	}

	var errors []error
	errors = append(errors, compareTerraformAndMarkdown(tfResources, readmeResources, "Resources")...)
	errors = append(errors, compareTerraformAndMarkdown(tfDataSources, readmeDataSources, "Data Sources")...)

	return errors
}
