ifndef DOCKERIMAGE
DOCKERIMAGE := quay.io/syncano/orion
endif

CURRENTPACKAGE := github.com/Syncano/orion
EXECNAME := orion

PATH := $(PATH):$(GOBIN):$(GOPATH)/bin
GOFILES = $(shell find . -mindepth 2 -type f -name '*.go' ! -path "./.*" ! -path "./assets/*" ! -path "./cmd/*" ! -path "./dev/*" ! -path "./vendor/*" ! -path "*/mocks/*" ! -path "*/proto/*")
GOPACKAGES = $(shell echo $(GOFILES) | xargs -n1 dirname | sort | uniq)

BUILDTIME = $(shell date +%Y-%m-%dT%H:%M)
GITSHA = $(shell git rev-parse --short HEAD)

LDFLAGS = -X github.com/Syncano/orion/pkg/version.GitSHA=$(GITSHA) \
	-X github.com/Syncano/orion/pkg/version.buildtimeStr=$(BUILDTIME)


.PHONY: help deps testdeps devdeps clean lint flint fmt test stest cov goconvey lint-in-docker test-in-docker generate-assets generate proto build build-static build-in-docker docker deploy-staging deploy-production encrypt decrypt start devserver run-server
.DEFAULT_GOAL := help
$(VERBOSE).SILENT:

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Install dep and sync vendored dependencies
	if ! which dep > /dev/null; then \
		echo "Installing dep"; \
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh; \
	fi
	dep ensure -v -vendor-only


testdeps: deps ## Install testing dependencies
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b $(GOPATH)/bin v1.12.2

devdeps: ## Install compile, testing and development dependencies
	if ! which protoc > /dev/null; then \
		echo "Installing proto"; \
		go get -u -v github.com/golang/protobuf/proto; \
	fi
	if ! which golangci-init > /dev/null; then \
		github.com/golangci/golangci-lint/cmd/golangci-lint; \
	fi
	go get -u -v github.com/smartystreets/goconvey
	go get -u -v github.com/gogo/protobuf/protoc-gen-gofast
	go get -u -v github.com/codegangsta/gin
	go get -u -v github.com/vektra/mockery/...
	go get -u -v github.com/jteeuwen/go-bindata/...
	go get -u -v golang.org/x/tools/cmd/goimports
	go get -u -v github.com/davecgh/go-spew/spew

require-%:
	if ! which ${*} > /dev/null; then \
		echo "! ${*} not installed"; \
		exit 1; \
	fi

clean: ## Cleanup repository
	go clean ./...
	rm -f build/$(EXECNAME)-static
	find deploy -name "*.unenc" -delete
	git clean -f

lint: require-golangci-lint ## Run fast lint checks
	echo "=== lint ==="
	golangci-lint run --fast

flint: require-golangci-lint ## Run full lint checks
	echo "=== full lint ==="
	golangci-lint run

fmt: ## Format code through goimports
	gofmt -s -w $(GOFILES)
	goimports -local $(CURRENTPACKAGE) -w $(GOFILES)

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

goconvey: require-goconvey ## Run goconvey test server
	goconvey -excludedDirs "vendor,dev,mocks,proto,assets,deploy,build" -timeout 5s -depth 2

lint-in-docker: require-docker-compose ## Run full lint in docker environment
	docker-compose run --no-deps --rm app make flint

test-in-docker: require-docker-compose ## Run full test suite in docker environment
	docker-compose run --rm app make build test

generate-assets: ## Generate assets with go-bindata
	go-bindata -nocompress -nometadata -nomemcopy -prefix assets -o assets/assets.go -pkg assets -ignore assets\.go assets/*
	
generate: generate-assets require-mockery ## Run go generate
	go generate $(GOPACKAGES)

proto: ## Run protobuf compiler on all .proto files
	for dir in $$(find . -name \*.proto -type f ! -path "./.*" ! -path "./vendor/*" -exec dirname {} \; | sort | uniq); do \
		protoc -I. \
			--gofast_out=plugins=grpc:$(GOPATH)/src \
			$$dir/*.proto; \
	done

proto-python: ## Run protobuf compiler on all .proto files and generate python files
	pip3 install grpcio-tools
	mkdir -p python
	for dir in $$(find . -name \*.proto -type f ! -path "./.*" ! -path "./vendor/*" -exec dirname {} \; | sort | uniq); do \
		python3 -m grpc_tools.protoc -I. --python_out=python --grpc_python_out=python $$dir/*.proto; \
	done

build: ## Build for current platform
	GIN_MODE=release go build -ldflags "$(LDFLAGS)" -o ./build/$(EXECNAME)

build-static: ## Build static version
	CGO_ENABLED=0 GIN_MODE=release go build -ldflags "-s $(LDFLAGS)" -a -installsuffix cgo -o ./build/$(EXECNAME)-static

build-in-docker: require-docker-compose ## Build static version in docker environment
	docker-compose run --no-deps --rm app make build-static

docker: require-docker ## Builds docker image for application (requires static version to be built first)
	docker build -t $(DOCKERIMAGE) build

deploy-staging: ## Deploy application to staging
	echo "=== deploying staging ==="
	kubectl config use-context k8s.syncano.rocks
	./deploy.sh staging stg-$(GITSHA) $(ARGS)

deploy-production: ## Deploy application to production
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

devserver: require-gin ## Run devserver
	DEBUG=1 FORCE_TERM=1 gin --port 8080 --bin build/$(EXECNAME) server

run-server: build ## Build and run server binary
	./build/$(EXECNAME) $(ARGS) server
