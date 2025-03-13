package markparsr

type TerraformDefinitionValidator struct {
	markdown  *MarkdownContent
	terraform *TerraformContent
}

func NewTerraformDefinitionValidator(markdown *MarkdownContent, terraform *TerraformContent) *TerraformDefinitionValidator {
	return &TerraformDefinitionValidator{
		markdown:  markdown,
		terraform: terraform,
	}
}

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
