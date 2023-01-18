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
	@go test -short -coverpkg=./... -coverprofile=${WORKSPACE}/gocoverage.out -count=1 ${GO_PKG_LIST} -json > ${WORKSPACE}/goreport.json


${WORKSPACE}/discovery_agent: dep
	@echo ${WORKSPACE}
	@export time=`date +%Y%m%d%H%M%S` && \
	export version=`cat version` && \
	@echo "makefile version"
	@cat version
	export commit_id=`cat commit_id` && \
	export sdk_version=`go list -m github.com/Axway/agent-sdk | awk '{print $$2}' | awk -F'-' '{print substr($$1, 2)}'` && \
	export GOOS=linux && \
	export CGO_ENABLED=0 && \
	export GOARCH=amd64 && \
	go build -tags static_all \
		-ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=$${time}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=$${version}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=$${commit_id}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.SDKBuildVersion=$${sdk_version}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=ApigeeDiscoveryAgent'" \
		-a -o ${WORKSPACE}/bin/apigee_discovery_agent ${WORKSPACE}/main.go

build: ${WORKSPACE}/discovery_agent	
	@echo "Build complete"

docker-build:
	@docker build -t apigee-discovery-agent:latest -f ${WORKSPACE}/docker/Dockerfile ../
	@echo "Docker build complete"
