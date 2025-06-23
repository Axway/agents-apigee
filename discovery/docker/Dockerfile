# Build image
# golang:1.24.4-alpine3.22 linux/amd64
FROM docker.io/golang@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder

ARG commit_id
ARG version
ARG sdk_version
ARG time
ARG CGO_ENABLED

RUN apk --no-cache update && \
  apk --no-cache add -U bash build-base git

ENV BASEPATH /go/src/github.com/Axway/agents-apigee
ENV AGENT ${BASEPATH}/discovery
ENV APP_USER axway

RUN mkdir -p ${BASEPATH}
WORKDIR ${BASEPATH}

# Copy necessary files
COPY . .

WORKDIR ${AGENT}

RUN export GOOS=linux && \
  export GOARCH=amd64 && \
  go build -tags static_all \
  -ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=${time}' \
  -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=${version}' \
  -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=${commit_id}' \
  -X 'github.com/Axway/agent-sdk/pkg/cmd.SDKBuildVersion=${sdk_version}' \
  -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=ApigeeDiscoveryAgent'" \
  -a -o ${AGENT}/bin/apigee_discovery_agent ${AGENT}/main.go

# Create non-root user
RUN addgroup -g 2500 ${APP_USER} && adduser -u 2500 -D -G ${APP_USER} ${APP_USER}
RUN chown -R ${APP_USER}:${APP_USER} ${AGENT}/bin/apigee_discovery_agent
USER ${APP_USER}

# alpine 3.22 linux/amd64
FROM docker.io/alpine@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715

ENV BASEPATH /go/src/github.com/Axway/agents-apigee
ENV AGENT ${BASEPATH}/discovery
ENV APP_USER axway

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder ${AGENT}/bin/apigee_discovery_agent /apigee_discovery_agent

RUN mkdir /keys /specs && \
  chown -R ${APP_USER} /keys /specs && \
  touch /apigee_discovery_agent.yml && \
  find / -perm /6000 -type f -exec chmod a-s {} \; || true

USER ${APP_USER}

VOLUME ["/keys", "/specs"]

HEALTHCHECK --retries=10 --start-period=30s --start-interval=30s CMD /apigee_discovery_agent --status || exit 1

ENTRYPOINT ["/apigee_discovery_agent"]