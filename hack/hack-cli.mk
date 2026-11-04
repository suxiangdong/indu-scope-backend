# Keep the CLI version aligned with the GoFrame version in go.mod.
GF_VERSION ?= v2.10.3
GF_BIN ?= $(if $(shell go env GOBIN),$(shell go env GOBIN),$(shell go env GOPATH)/bin)/gf

# Install/update the project GoFrame CLI tool.
.PHONY: cli
cli:
	@set -e; \
		echo "go install github.com/gogf/gf/cmd/gf/v2@$(GF_VERSION)"; \
		go install github.com/gogf/gf/cmd/gf/v2@$(GF_VERSION); \
		echo "GoFrame CLI installed successfully!"

# Check and install the project CLI tool.
.PHONY: cli.install
cli.install:
	@set -e; \
		cli_version=$$("$(GF_BIN)" -v 2>/dev/null | sed -n '1p' || true); \
		if [ "$$cli_version" != "$(GF_VERSION)" ]; then \
			echo "GoFrame CLI is not installed, installing $(GF_VERSION)..."; \
			$(MAKE) cli; \
		fi
