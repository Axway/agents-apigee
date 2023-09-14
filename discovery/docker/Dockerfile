# Build image
# golang:1.21.1-alpine3.18 linux/amd64 
FROM docker.io/golang@sha256:c5c5fda71656f28e49ac9c5416b3643eaa6a108a8093151d6d1afc9463be8e33 AS builder

ENV GOFLAGS "-mod=mod"
ENV GOWORK "off"

RUN apk --no-cache update && \
  apk --no-cache add -U bash build-base git

ENV BASEPATH=/go/src/github.com/Axway/agents-apigee
ENV AGENT=${BASEPATH}/discovery

RUN mkdir -p ${BASEPATH}

WORKDIR ${BASEPATH}

# Copy necessary files
COPY . .

WORKDIR ${AGENT}

RUN export time=`date +%Y%m%d%H%M%S` && \
  export commit_id=`git rev-parse --short HEAD` && \
  export version=`git tag -l --sort='version:refname' | grep -Eo '[0-9]{1,}\.[0-9]{1,}\.[0-9]{1,3}$' | tail -1` && \
  export sdk_version=`go list -m github.com/Axway/agent-sdk | awk '{print $2}' | awk -F'-' '{print substr($1, 2)}'` && \
  export GOOS=linux && \
  export CGO_ENABLED=0 && \
  export GOARCH=amd64 && \
  go build -tags static_all \
  -ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=${time}' \
  -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=${version}' \
  -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=${commit_id}' \
  -X 'github.com/Axway/agent-sdk/pkg/cmd.SDKBuildVersion=${sdk_version}' \
  -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=ApigeeDiscoveryAgent'" \
  -a -o ${AGENT}/bin/apigee_discovery_agent ${AGENT}/main.go

# Create non-root user
RUN addgroup -g 2500 axway && adduser -u 2500 -D -G axway axway
RUN chown -R axway:axway /go/src/github.com/Axway/agents-apigee/discovery/bin/apigee_discovery_agent
USER axway

# alpine 3.18.3
FROM docker.io/alpine@sha256:c5c5fda71656f28e49ac9c5416b3643eaa6a108a8093151d6d1afc9463be8e33

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/Axway/agents-apigee/discovery/bin/apigee_discovery_agent /apigee_discovery_agent

RUN mkdir /keys && \
  chown -R axway /keys && \
  apk --no-cache add openssl libssl3 libcrypto3 musl musl-utils libc6-compat busybox curl && \
  touch /apigee_discovery_agent.yml && \
  find / -perm /6000 -type f -exec chmod a-s {} \; || true

USER axway

VOLUME ["/keys"]

HEALTHCHECK --retries=1 CMD /apigee_discovery_agent --status || exit 1

ENTRYPOINT ["/apigee_discovery_agent"]