# Build image
# golang:1.19.0-alpine3.16 linux/amd64
FROM docker.io/golang@sha256:70df3b8f9f099da7f60f0b32480015165e3d0b51bfacf9e255b59f3dd6bd2828 as builder

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
  export version=`cat version` && \
  export commit_id=`git rev-parse --short HEAD` && \
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

# alpine 3.15.0
FROM docker.io/alpine@sha256:e7d88de73db3d3fd9b2d63aa7f447a10fd0220b7cbf39803c803f2af9ba256b3

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/Axway/agents-apigee/discovery/bin/apigee_discovery_agent /apigee_discovery_agent

RUN mkdir /keys && \
  chown -R axway /keys && \
  apk --no-cache add openssl libssl1.1 libcrypto1.1 musl musl-utils libc6-compat busybox curl && \
  touch /apigee_discovery_agent.yml && \
  find / -perm /6000 -type f -exec chmod a-s {} \; || true

USER axway

VOLUME ["/keys"]

HEALTHCHECK --retries=1 CMD curl --fail http://localhost:${STATUS_PORT:-8989}/status || exit 1

ENTRYPOINT ["/apigee_discovery_agent"]