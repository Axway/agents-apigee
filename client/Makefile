.PHONY: all dep test build

WORKSPACE ?= $$(pwd)
GO_PKG_LIST := $(shell go list ./...)

export GOFLAGS := -mod=mod
export GOWORK := off
export GOPRIVATE := git.ecd.axway.org

all: clean package
	@echo "Done"

clean:
	@rm -rf ./bin/
	@mkdir -p ./bin
	@echo "Clean complete"

dep-check:
	@go mod verify

dep:
	@echo "Resolving go package dependencies"
	@go mod tidy
	@echo "Package dependencies completed"

dep-branch:
	@make sdk=`git branch --show-current` dep-version

dep-version:
	@export version=$(sdk) && make update-sdk && make dep

dep-sdk: 
	@make sdk=main dep-version

update-sdk:
	@echo "Updating SDK dependencies"
	@export GOFLAGS="" && go mod edit -require "github.com/Axway/agent-sdk@${version}"

test: dep
	@go vet ${GO_PKG_LIST}
	@go test -race -v -short -coverprofile=${WORKSPACE}/gocoverage.out -count=1 ${GO_PKG_LIST}

test-sonar: dep
	@go vet ${GO_PKG_LIST}
	@go test -v -short -coverpkg=./... -coverprofile=./gocoverage.out -count=1 ${GO_PKG_LIST} -json > ./goreport.json
	
sonar: test-sonar
	./build/scripts/sonar.sh $(sonarHost)

apigee-generate: ## generate apigee resources
	docker run --rm -v $(shell pwd)/specs:/codegen/specs -v $(shell pwd)/scripts:/codegen/scripts -v $(shell pwd)/pkg/apigee/models:/codegen/output  -e USERID=$(shell id -u) -e GROUPID=$(shell id -g) -w /codegen/scripts --entrypoint ./apigee_generate.sh ampc-beano-docker-snapshot-phx.artifactory-phx.ecd.axway.int/beano-alpine-codegen:latest
