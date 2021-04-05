.PHONY: all dep test lint build 

WORKSPACE ?= $$(pwd)

GO_PKG_LIST := $(shell go list ./... | grep -v /vendor/)

lint:
	@golint -set_exit_status ${GO_PKG_LIST}

dep:
	@echo "Resolving go package dependencies"
	@go mod tidy
	@go mod vendor
	@echo "Package dependencies completed"

update-sdk:
	@echo "Updating SDK dependencies"
	@export GOFLAGS="" && go get "github.com/Axway/agent-sdk@main"


${WORKSPACE}/apigee_traceability_agent: dep
	@export time=`date +%Y%m%d%H%M%S` && \
	export version=`cat version` && \
	export commit_id=`git rev-parse --short HEAD` && \
	go build -tags static_all \
		-ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=$${time}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=$${version}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=$${commit_id}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=APIGEETraceabilityAgent'" \
		-a -o ${WORKSPACE}/bin/apigee_traceability_agent ${WORKSPACE}/main.go

build:${WORKSPACE}/apigee_traceability_agent
	@echo "Build complete"

docker-build:
	@docker build -t apigee-traceability-agent:latest -f ${WORKSPACE}/docker/Dockerfile .
	@echo "Docker build complete"