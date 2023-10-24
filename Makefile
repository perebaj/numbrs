SHELL = /bin/bash

# TOOLS VERSIONS
GO_VERSION=1.21.3
GOLANGCI_LINT_VERSION=v1.54.0

# configuration/aliases
version=$(shell git rev-parse --short HEAD)
base_image=perebaj/numbrs
image=$(base_image):$(version)
devimage=numbrs-dev
# To avoid downloading deps everytime it runs on containers
gopkg=$(devimage)-gopkg
gocache=$(devimage)-gocache
devrun=docker run $(devrunopts) --rm \
	-v `pwd`:/app \
	-v $(gopkg):/go/pkg \
	-v $(gocache):/root/.cache/go-build \
	$(devimage)

covreport ?= coverage.txt

all: lint test image

## run isolated tests
.PHONY: test
test:
	go test -run="$(testcase)" -cover -race -shuffle on ./... -v

## Format go code
.PHONY: fmt
fmt:
	goimports -w .

## builds the service
.PHONY: service
service:
	go build -o ./cmd/numbrs/numbrs ./cmd/numbrs

## runs the service locally
.PHONY: run
run: service
	./cmd/numbrs/numbrs

## lint the whole project
.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## generates coverage report
.PHONY: test/coverage
test/coverage:
	go test -count=1 -coverprofile=$(covreport) ./...

## generates coverage report and shows it on the browser locally
.PHONY: test/coverage/show
test/coverage/show: test/coverage
	go tool cover -html=$(covreport)

## Build the service image
.PHONY: image
image:
	docker build . \
		--build-arg GO_VERSION=$(GO_VERSION) \
		-t $(image)

## Build a production ready container image and run it locally for testing.
.PHONY: image/run
image/run: image
	docker run --rm -ti \
		-v $(gopkg):/go/pkg \
		$(image)

## Publish the service image
.PHONY: image/publish
image/publish: image
	docker push $(image)

## Create the dev container image
.PHONY: dev/image
dev/image:
	docker build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
		-t $(devimage) \
		-f Dockerfile.dev \
		.

## Create a shell inside the dev container
.PHONY: dev
dev: devrunopts=-ti
dev: dev/image
	$(devrun)

## run a make target inside the dev container.
dev/%: dev/image
	$(devrun) make ${*}

## Display help for all targets
.PHONY: help
help:
	@awk '/^.PHONY: / { \
		msg = match(lastLine, /^## /); \
			if (msg) { \
				cmd = substr($$0, 9, 100); \
				msg = substr(lastLine, 4, 1000); \
				printf "  ${GREEN}%-30s${RESET} %s\n", cmd, msg; \
			} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
