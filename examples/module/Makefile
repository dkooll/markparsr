.PHONY: all install-tools fmt document table

all: install-tools fmt docs

install-tools:
	go install github.com/terraform-docs/terraform-docs@latest

document:
	@echo "Generating document-style documentation for root and modules..."
	terraform-docs markdown document . --output-file README.md --output-mode inject --hide modules
	for dir in modules/*; do \
		if [ -d "$$dir" ]; then \
			echo "Processing $$dir..."; \
			terraform-docs markdown document "$$dir" --output-file "$$dir/README.md" --output-mode inject --hide modules || echo "Skipped: $$dir"; \
		fi \
	done

table:
	@echo "Generating table-style documentation for root and modules..."
	terraform-docs markdown table . --output-file README.md --output-mode inject --hide modules
	for dir in modules/*; do \
		if [ -d "$$dir" ]; then \
			echo "Processing $$dir..."; \
			terraform-docs markdown table "$$dir" --output-file "$$dir/README.md" --output-mode inject --hide modules || echo "Skipped: $$dir"; \
		fi \
	done

fmt:
	terraform fmt -recursive
