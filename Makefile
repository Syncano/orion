ifndef DOCKERIMAGE
DOCKERIMAGE := quay.io/syncano/orion
endif

CURRENTPACKAGE := github.com/Syncano/orion
EXECNAME := orion

PATH := $(PATH):$(GOPATH)/bin
GOFILES=$(shell find . -mindepth 2 -type f -name '*.go' ! -path "./.*" ! -path "./dev/*" ! -path "*/proto/*")
GOTESTPACKAGES = $(shell find . -mindepth 2 -type f -name '*.go' ! -path "./.*" ! -path "./internal/*" ! -path "./dev/*" ! -path "*/mocks/*" ! -path "*/proto/*" | xargs -n1 dirname | sort | uniq)

BUILDTIME = $(shell date +%Y-%m-%dT%H:%M)
GITSHA = $(shell git rev-parse --short HEAD)

LDFLAGS = -s -w \
	-X github.com/Syncano/orion/pkg/version.GitSHA=$(GITSHA) \
	-X github.com/Syncano/orion/pkg/version.buildtimeStr=$(BUILDTIME)


.PHONY: help clean lint fmt test stest cov goconvey lint-in-docker test-in-docker proto build build-in-docker docker deploy-staging deploy-production encrypt decrypt start devserver run-server
.DEFAULT_GOAL := help
$(VERBOSE).SILENT:

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

require-%:
	if ! which ${*} > /dev/null; then \
		echo "! ${*} not installed"; \
		exit 1; \
	fi

clean: ## Cleanup repository
	go clean ./...
	rm -f build/$(EXECNAME)
	find deploy -name "*.unenc" -delete
	git clean -f

lint: ## Run lint checks
	echo "=== lint ==="
	if ! which golangci-lint > /dev/null; then \
		echo "Installing golangci-lint"; \
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $$(go env GOPATH)/bin v1.20.0; \
	fi
	golangci-lint run $(ARGS)

fmt: ## Format code through goimports
	gofmt -s -w $(GOFILES)
	go run golang.org/x/tools/cmd/goimports -local $(CURRENTPACKAGE) -w $(GOFILES)

test: ## Run unit with race check and create coverage profile
	echo "=== unit test ==="
	echo "mode: atomic" > coverage-all.out
	$(foreach pkg,$(GOPACKAGES),\
		go test -timeout 5s -race -coverprofile=coverage.out -covermode=atomic $(ARGS) $(pkg) || exit;\
		tail -n +2 coverage.out >> coverage-all.out 2>/dev/null;)

stest: ## Run only short tests (unit tests) without race check
	echo "=== short test ==="
	go test -timeout 5s -short $(ARGS) $(GOPACKAGES)

cov: ## Show per function coverage generated by test
	echo "=== coverage ==="
	go tool cover -func=coverage-all.out

goconvey: ## Run goconvey test server
	go run github.com/smartystreets/goconvey -excludedDirs "dev,internal,mocks,proto,assets,deploy,build" -timeout 5s -depth 2

lint-in-docker: require-docker-compose ## Run lint in docker environment
	docker-compose run --no-deps --rm app make lint

test-in-docker: require-docker-compose ## Run full test suite in docker environment
	docker-compose run --rm app make build test

proto: ## Run protobuf compiler on all .proto files
	for dir in $$(find . -name \*.proto -type f ! -path "./.*" -exec dirname {} \; | sort | uniq); do \
		protoc -I. \
			--gofast_out=plugins=grpc:$(GOPATH)/src \
			$$dir/*.proto; \
	done

proto-python: ## Run protobuf compiler on all .proto files and generate python files
	pip3 install grpcio-tools
	mkdir -p python
	for dir in $$(find . -name \*.proto -type f ! -path "./.*" -exec dirname {} \; | sort | uniq); do \
		python3 -m grpc_tools.protoc -I. --python_out=python --grpc_python_out=python $$dir/*.proto; \
	done

build: ## Build
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./build/$(EXECNAME)

build-in-docker: require-docker-compose ## Build in docker environment
	docker-compose run --no-deps --rm app make build build-wrapper

docker: require-docker ## Builds docker image for application (requires static version to be built first)
	docker build -t $(DOCKERIMAGE) build

deploy-staging: require-kubectl ## Deploy application to staging
	echo "=== deploying staging ==="
	kubectl config use-context k8s.syncano.rocks
	./deploy.sh staging stg-$(GITSHA) $(ARGS)

deploy-production: require-kubectl ## Deploy application to production
	echo "=== deploying us1 ==="
	kubectl config use-context k8s.syncano.io
	./deploy.sh us1 prd-$(GITSHA) $(ARGS)

	echo "=== deploying eu1 ==="
	kubectl config use-context gke_pioner-syncano-prod-9cfb_europe-west1_syncano-eu1
	./deploy.sh eu1 prd-$(GITSHA) --skip-push

encrypt: ## Encrypt unencrypted files (for secrets).
	find deploy -name "*.unenc" -exec sh -c 'gpg --batch --yes --passphrase "$(ORION_VAULT_PASS)" --symmetric --cipher-algo AES256 -o "$${1%.unenc}.gpg" "$$1"' _ {} \;

decrypt: ## Decrypt files.
	find deploy -name "*.gpg" -exec sh -c 'gpg --batch --yes --passphrase "$(ORION_VAULT_PASS)" --decrypt -o "$${1%.gpg}.unenc" "$$1"' _ {} \;

start: require-docker-compose ## Run docker-compose of an app.
	docker-compose -f build/docker-compose.yml up

devserver: ## Run devserver
	DEBUG=1 FORCE_TERM=1 go run github.com/codegangsta/gin --port 8080 --bin build/$(EXECNAME) server

run-server: build ## Build and run server binary
	./build/$(EXECNAME) $(ARGS) server
