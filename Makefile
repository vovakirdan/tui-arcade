.PHONY: all 

GO ?= go

GOBIN := $(shell $(GO) env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell $(GO) env GOPATH)/bin
endif

GOLANGCI_LINT := $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION := v1.62.2

build:
	@echo ">> Building"
	$(GO) build -o arcade ./cmd/arcade

fmt:
	@echo ">> Formatting"
	$(GO) fmt ./...

lint: $(GOLANGCI_LINT)
	@echo ">> Running linters"
	$(GOLANGCI_LINT) run --config .golangci.yaml

apply-cfg:
	@echo ">> Applying config"
	cp internal/config/defaults/flappy.yaml ~/.arcade/configs/flappy.yaml
	cp internal/config/defaults/dino.yaml ~/.arcade/configs/dino.yaml

clean:
	@echo ">> Cleaning"
	rm -f arcade

test:
	@echo ">> Running tests"
	go test ./...