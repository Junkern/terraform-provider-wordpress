HOSTNAME=registry.terraform.io
NAMESPACE=junkern
NAME=wordpress
BINARY=terraform-provider-${NAME}
VERSION=9.9.9
OS_ARCH=$(shell go env GOHOSTOS)_$(shell go env GOHOSTARCH)

GOLANGCI_VERSION = 2.12.2

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

local-build:
	go build -o ${BINARY}

local-install: local-build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint

bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

golangci-lint: bin/golangci-lint
	@bin/golangci-lint run ./...