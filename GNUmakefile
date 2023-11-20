HOSTNAME=registry.terraform.io
NAMESPACE=junkern
NAME=wordpress
BINARY=terraform-provider-${NAME}
VERSION=9.9.9
OS_ARCH=$(shell go env GOHOSTOS)_$(shell go env GOHOSTARCH)

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