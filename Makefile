.DEFAULT_GOAL := help

GOLANGCI_VERSION := 1.53.2

.PHONY: install-lint
install-lint: ## Installs the required version of golangci-lint. *Careful*, this will override your existing install if you have one!
	@# golangci-lint team doesnt recommend install the tool via "go get"
	@# https://golangci-lint.run/usage/install/
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$${GOBIN:-$(shell go env GOPATH)/bin}" v$(GOLANGCI_VERSION)

.PHONY: ensure-lint
ensure-lint: ## Ensure that the linting binary is the correct version
	@golangci-lint --version | grep $(GOLANGCI_VERSION) > /dev/null || echo "Please install version $(GOLANGCI_VERSION) of \`golangci-lint\` by running \`make install-lint\`"

.PHONY: lint
lint: ensure-lint ## Lint the application
	golangci-lint run --timeout=3m ./...;

.PHONY: test
test: ## Run the tests
	go test -failfast -short -race ./...

.PHONY: help
help: ## Prints this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\033[36m"; } /^[a-zA-Z_-]+:.*?##/ { printf "  %-30s %s\n", $$1, $$2; }' $(MAKEFILE_LIST) | sort