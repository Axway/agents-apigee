# Build image
# golang:1.16.2-alpine3.13 linux/amd64
FROM docker.io/golang@sha256:12d5f94cd4d2840e538e82e26a5dfddf711b30cc98a9f6e01bcf65d7aaf7ccd8 as builder

RUN mkdir -p /go/src/github.com/Axway/agents-apigee/traceability

WORKDIR /go/src/github.com/Axway/agents-apigee/traceability

# Copy necessary files
COPY . . 

RUN apk --no-cache update && \
  apk --no-cache add -U bash build-base git

RUN make build

# Create non-root user
RUN addgroup -g 2500 axway && adduser -u 2500 -D -G axway axway
RUN chown -R axway:axway /go/src/github.com/Axway/agents-apigee/traceability/bin/apigee_traceability_agent
USER axway

# alpine 3.13.2
FROM docker.io/alpine@sha256:4661fb57f7890b9145907a1fe2555091d333ff3d28db86c3bb906f6a2be93c87

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/Axway/agents-apigee/traceability/apigee_traceability_agent.yml /apigee_traceability_agent.yml
COPY --from=builder /go/src/github.com/Axway/agents-apigee/traceability/bin/apigee_traceability_agent /apigee_traceability_agent

RUN apk --no-cache add openssl libssl1.1 libcrypto1.1 musl musl-utils libc6-compat busybox curl && \
  mkdir /keys /data /events && \
  chmod go-w /apigee_traceability_agent.yml && \
  chown -R axway:2500 /keys /data /events /apigee_traceability_agent.yml


USER axway

VOLUME ["/keys"]

HEALTHCHECK --retries=1 CMD curl --fail http://localhost:${STATUS_PORT:-8989}/status || exit 1

ENTRYPOINT ["/apigee_traceability_agent"]