
WORKSPACE := ${PWD}
DIRS := ${WORKSPACE}/client/pkg/... ${WORKSPACE}/discovery/... ${WORKSPACE}/traceability/...
GO_PKG_LIST := $(shell go list ${DIRS})

# export GOFLAGS := -mod=mod
# export GOWORK := ${PWD}/go.work
# export GOPRIVATE := git.ecd.axway.org

dep:
	@$(MAKE) -C client dep
	@$(MAKE) -C discovery dep
	@$(MAKE) -C traceability dep


dep-branch:
	@$(MAKE) -C client dep-branch
	@$(MAKE) -C discovery dep-branch
	@$(MAKE) -C traceability dep-branch

dep-version:
	@$(MAKE) -C client dep-version sdk=$(sdk)
	@$(MAKE) -C discovery dep-version sdk=$(sdk)
	@$(MAKE) -C traceability dep-version sdk=$(sdk)

dep-sdk: 
	@$(MAKE) -C client dep-sdk
	@$(MAKE) -C discovery dep-sdk
	@$(MAKE) -C traceability dep-sdk

test-sonar:
	# @go vet ${GO_PKG_LIST}
	@echo "${PWD}"
	@echo "${WORKSPACE}"
	@go work use .
	@go test -short -coverpkg=${GO_PKG_LIST} -coverprofile=${WORKSPACE}/gocoverage.out -count=1 ${GO_PKG_LIST} -json > ${WORKSPACE}/goreport.json

sonar: test-sonar
	./sonar.sh $(sonarHost)

