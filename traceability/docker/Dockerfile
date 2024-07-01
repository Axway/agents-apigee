# Build image
# golang:1.22.4-alpine3.20 linux/amd64
FROM docker.io/golang@sha256:ace6cc3fe58d0c7b12303c57afe6d6724851152df55e08057b43990b927ad5e8 AS builder

ARG commit_id
ARG version
ARG sdk_version
ARG time
ARG CGO_ENABLED

RUN apk --no-cache update && \
  apk --no-cache add -U bash build-base git

ENV BASEPATH /go/src/github.com/Axway/agents-apigee
ENV AGENT ${BASEPATH}/traceability
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
  -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=ApigeTraceabilityAgent'" \
  -a -o ${AGENT}/bin/apigee_traceability_agent ${AGENT}/main.go

# Create non-root user
RUN addgroup -g 2500 ${APP_USER} && adduser -u 2500 -D -G ${APP_USER} ${APP_USER}
RUN chown -R ${APP_USER}:${APP_USER} ${AGENT}/bin/apigee_traceability_agent
USER ${APP_USER}

# alpine 3.20.1 linux/amd64
FROM docker.io/alpine@sha256:b89d9c93e9ed3597455c90a0b88a8bbb5cb7188438f70953fede212a0c4394e0

ENV BASEPATH /go/src/github.com/Axway/agents-apigee
ENV AGENT ${BASEPATH}/traceability
ENV APP_USER axway

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder ${AGENT}/bin/apigee_traceability_agent /apigee_traceability_agent
COPY --from=builder ${AGENT}/apigee_traceability_agent.yml /apigee_traceability_agent.yml

RUN mkdir /keys /data && \
  chown -R ${APP_USER} /keys /data /apigee_traceability_agent.yml && \
  chmod go-w /apigee_traceability_agent.yml

RUN find / -perm /6000 -type f -exec chmod a-s {} \; || true

RUN chgrp -R 0 /apigee_traceability_agent /data && chmod -R g=u /apigee_traceability_agent /data && chown -R axway /apigee_traceability_agent /data
RUN chown 0 /apigee_traceability_agent.yml && chmod go-w /apigee_traceability_agent.yml

USER ${APP_USER}

VOLUME ["/keys", "/data"]

HEALTHCHECK --retries=1 CMD /apigee_traceability_agent --status || exit 1

ENTRYPOINT ["/apigee_traceability_agent"]