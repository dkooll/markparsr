# markparsr [![Go Reference](https://pkg.go.dev/badge/github.com/dkooll/markparsr.svg)](https://pkg.go.dev/github.com/dkooll/markparsr)

A terraform documentation validation tool that keeps README files aligned with their module sources.

Ensures your module docs stay accurate, highlights drift automatically, and provides detailed reporting for reliable infrastructure documentation.

## Why markparsr?

Terraform modules evolve rapidly and documentation often lags behind—missing sections, outdated variable descriptions, or stale resource lists create confusion.

Manual auditing is tedious and error-prone.

`markparsr helps you:`

Validate docs against Terraform source before shipping changes.

Keep documentation aligned with Terraform definitions.

Run lightweight checks in CI/CD for every module.

Support custom sections, provider prefixes, and file requirements.

Automate documentation hygiene across teams and repositories.

## Installation

`go get github.com/dkooll/markparsr`

## Usage

See the [examples/](examples/) directory for sample Terraform modules and validator tests.

Run the Go tests inside `examples/usage/` to validate the bundled example module.

## Features

`README Section Validation`

Enforces Terraform-docs sections (Requirements, Providers, Inputs, Outputs, Resources).

Detects missing or misspelled headings with typo-friendly matching.

Extracts items even when headings disappear by leveraging anchors.

`HCL ↔ README Consistency`

Compares documented variables and outputs with those declared in HCL.

Verifies resources and data sources referenced in the README actually exist in code.

Supports provider prefix configuration for custom naming schemes.

`File & URL Checks`

Ensures key module files (README, variables.tf, outputs.tf, terraform.tf) are present and non-empty.

Validates URLs in the README respond successfully.

`Flexible Configuration`

Functional options for additional sections, extra files, provider prefixes, and README paths.

Environment variable overrides for CI/CD (`README_PATH`, `MODULE_PATH`, `FORMAT`, `VERBOSE`).

Lightweight output suitable for Go test integration and automation.

## Configuration

`Functional Options`

`WithFormat(format)`: Force the markdown format (defaults to `document`).

`WithAdditionalSections(sections...)`: Require extra documentation sections.

`WithAdditionalFiles(files...)`: Ensure additional files exist beside Terraform defaults.

`WithRelativeReadmePath(path)`: Point to the README when it is not in the module root.

`WithProviderPrefixes(prefixes...)`: Recognize custom resource prefixes.

`Environment Variables`

`README_PATH`: Absolute README path when not passed via options.

`MODULE_PATH`: Module root directory (defaults to the README directory).

`FORMAT`: Set to `document`; other values fall back to document mode with a warning.

`VERBOSE`: When `true`, prints diagnostic information.

### Notes

markparsr assumes Terraform-docs style READMEs with H2/H3 headings and anchor links.

Provider prefixes help resource detection across custom modules and registries.

Run validators in CI to prevent documentation drift before merging.

## Contributors

We welcome contributions from the community! Whether it's reporting a bug, suggesting a new feature, or submitting a pull request, your input is highly valued. <br><br>

<a href="https://github.com/dkooll/markparsr/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=dkooll/markparsr" />
</a>
