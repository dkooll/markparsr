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

	readmeResources, readmeDataSources, mdErr := tdv.markdown.ExtractResourcesAndDataSources()

	collector := &ErrorCollector{}
	if len(tfResources)+len(tfDataSources) > 0 {
		if mdErr != nil {
			collector.Add(mdErr)
			return collector.Errors()
		}
	}

	if tdv.markdown.HasSection("Resources") || len(readmeResources) > 0 || len(readmeDataSources) > 0 {
		collector.AddMany(compareTerraformAndMarkdown(tfResources, readmeResources, "Resources"))
		collector.AddMany(compareTerraformAndMarkdown(tfDataSources, readmeDataSources, "Data Sources"))
	}
	return collector.Errors()
}
