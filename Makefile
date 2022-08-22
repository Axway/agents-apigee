
# export GOFLAGS := -mod=readonly
# export GOWORK := $$(pwd)/go.work

WORKSPACE := $$(pwd)
GO_PKG_LIST := $(shell go list ./client/pkg/... ./discovery/... ./traceability/...)

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
	ls -laR
	echo "wkspc"
	echo ${WORKSPACE}
	echo "gopkglist"
	# @export GOFLAGS="-mod=readonly" && \
	echo "${GO_PKG_LIST}"
	@echo "THREE"
	go test -v -short -coverpkg=./... -coverprofile=${WORKSPACE}/gocoverage.out -count=1 ${GO_PKG_LIST} -json > ${WORKSPACE}/goreport.json && \
	echo "THERE"

sonar: test-sonar
	@GO_PKG_LIST=$(shell go list ./client/pkg/... ./discovery/... ./traceability/...)
	ls -laR
	./sonar.sh $(sonarHost)

