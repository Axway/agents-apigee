# Build image
# golang:1.19.8-alpine3.17 linux/amd64
FROM docker.io/golang@sha256:841c160ed35923d96c95c52403c4e6db5decd9cbce034aa851e412ade5d4b74f as builder

ENV GOFLAGS "-mod=mod"
ENV GOWORK "off"

RUN apk --no-cache update && \
  apk --no-cache add -U bash build-base git

ENV BASEPATH=/go/src/github.com/Axway/agents-apigee
ENV AGENT=${BASEPATH}/traceability

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
  -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=ApigeTraceabilityAgent'" \
  -a -o ${AGENT}/bin/apigee_traceability_agent ${AGENT}/main.go

# Create non-root user
RUN addgroup -g 2500 axway && adduser -u 2500 -D -G axway axway
RUN chown -R axway:axway /go/src/github.com/Axway/agents-apigee/traceability/bin/apigee_traceability_agent
USER axway

# alpine 3.17.0
FROM docker.io/alpine@sha256:c0d488a800e4127c334ad20d61d7bc21b4097540327217dfab52262adc02380c

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/Axway/agents-apigee/traceability/apigee_traceability_agent.yml /apigee_traceability_agent.yml
COPY --from=builder /go/src/github.com/Axway/agents-apigee/traceability/bin/apigee_traceability_agent /apigee_traceability_agent

RUN apk --no-cache add openssl libssl1.1 libcrypto1.1 musl musl-utils libc6-compat busybox curl && \
  mkdir /keys /data && \
  chown -R axway /keys /data /apigee_traceability_agent.yml && \
  chmod go-w /apigee_traceability_agent.yml

RUN find / -perm /6000 -type f -exec chmod a-s {} \; || true

RUN chgrp -R 0 /apigee_traceability_agent /data && chmod -R g=u /apigee_traceability_agent /data && chown -R axway /apigee_traceability_agent /data
RUN chown 0 /apigee_traceability_agent.yml && chmod go-w /apigee_traceability_agent.yml

USER axway

VOLUME ["/keys", "/data"]

HEALTHCHECK --retries=1 CMD curl --fail http://localhost:${STATUS_PORT:-8989}/status || exit 1

ENTRYPOINT ["/apigee_traceability_agent"]