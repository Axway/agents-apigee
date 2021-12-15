.PHONY: all dep test lint build

WORKSPACE ?= $$(pwd)

GO_PKG_LIST := $(shell go list ./...)

export GOFLAGS := -mod=mod
export GOPRIVATE := git.ecd.axway.org

lint:
	@golint -set_exit_status ${GO_PKG_LIST}

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
	@export GOFLAGS="" && go get "github.com/Axway/agent-sdk@${version}"

test: dep
	@go vet ${GO_PKG_LIST}
	@go test -race -v -short -coverprofile=${WORKSPACE}/gocoverage.out -count=1 ${GO_PKG_LIST}

test-sonar: dep
	@go vet ${GO_PKG_LIST}
	@go test -short -coverpkg=./... -coverprofile=${WORKSPACE}/gocoverage.out -count=1 ${GO_PKG_LIST} -json > ${WORKSPACE}/goreport.json

apigee-generate: ## generate apigee resources
	./scripts/apigee_generate.sh