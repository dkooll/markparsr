/*
Package markparsr validates that terraform module documentation is complete and accurate.

markparsr ensures there's consistency between your Terraform code and README documentation,
helping maintain documentation quality as your module evolves. The package analyzes both
HCL files and markdown to identify gaps or inconsistencies.

Key validations:
  - Resources/data sources match between code and docs
  - Variables and outputs are properly documented
  - Required documentation sections are present
  - Referenced URLs are accessible
  - Required files exist and aren't empty

Usage:
	validator, _ := markparsr.NewReadmeValidator("path/to/README.md")
	errors := validator.Validate()
	// Handle errors as needed

The package assumes a standard Terraform module structure with README.md,
variables.tf, outputs.tf, and other standard files in the same directory.
*/
package markparsr
