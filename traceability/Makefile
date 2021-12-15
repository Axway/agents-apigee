.PHONY: all dep test lint build 

WORKSPACE ?= $$(pwd)

GO_PKG_LIST := $(shell go list ./... | grep -v /vendor/)

export GOFLAGS := -mod=mod
export GOPRIVATE := git.ecd.axway.org

lint:
	@golint -set_exit_status ${GO_PKG_LIST}

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

${WORKSPACE}/apigee_traceability_agent: dep
	@export time=`date +%Y%m%d%H%M%S` && \
	export version=`cat version` && \
	export commit_id=`cat commit_id` && \
	export sdk_version=`go list -m github.com/Axway/agent-sdk | awk '{print $$2}' | awk -F'-' '{print substr($$1, 2)}'` && \
	export GOOS=linux && \
	export CGO_ENABLED=0 && \
	export GOARCH=amd64 && \
	go build -tags static_all \
		-ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=$${time}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=$${version}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=$${commit_id}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=ApigeeTraceabilityAgent'" \
		-a -o ${WORKSPACE}/bin/apigee_traceability_agent ${WORKSPACE}/main.go

build:${WORKSPACE}/apigee_traceability_agent
	@echo "Build complete"

docker-build:
	@docker build -t apigee-traceability-agent:latest -f ${WORKSPACE}/docker/Dockerfile .
	@echo "Docker build complete"