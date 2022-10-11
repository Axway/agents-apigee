# Apigee Traceability Agent

The Traceability agent finds logs from consumed Apigee proxies and sends the traffic data to Amplify Central

## Build and run

The following make targets are available

| Target       | Description                                                    | Output(s)                        |
| ------------ | -------------------------------------------------------------- | -------------------------------- |
| lint         | runs go lint against all source files                          | linter results                   |
| dep          | downloads all dependencies needed to build the discovery agent | /vendor                          |
| test         | runs go test against all test files int he repo                | test results                     |
| update-sdk   | pulls the latest changes to main on the SDK repo               |                                  |
| build        | builds the binary traceability agent                           | bin/apigee_traceability_agent    |
| docker-build | builds the traceability agent in a docker container            | apigee-traceability-agent:latest |

### Build (Docker)

```
make docker-build
```

### Run (Docker)

```
docker run --env-file env_vars  -v `pwd`/data:/data -v `pwd`/keys:/keys apigee-traceability-agent:latest
```

### Build (Windows)

* Build the agent using the following command

```shell
go build -tags static_all \
    -ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=$${time}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=$${version}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=$${commit_id}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=APIGEETraceabilityAgent'" \
    -a -o ./bin/apigee_traceability_agent.exe ./main.go
```

### Run (Windows)

* After a successful build, you should see the executable under the bin folder.   And you can execute it using the following command

```shell
./apigee_traceability_agent.exe --envFile env_vars
```

## Traceability agent variables

| Environment Variable       | Description                                      | Default (if applicable)           |
| -------------------------- | ------------------------------------------------ | --------------------------------- |
| APIGEE_URL                 | The base Apigee URL for this agent to connect to | https://api.enterprise.apigee.com |
| APIGEE_APIVERSION          | The version of the API for the agent to use      | v1                                |
| APIGEE_ORGANIZATION        | The Apigee organization name                     |                                   |
| APIGEE_AUTH_USERNAME       | The Apigee account username/email address        |                                   |
| APIGEE_AUTH_PASSWORD       | The Apigee account password                      |                                   |
| APIGEE_AUTH_URL            | The IDP URL                                      | https://login.apigee.com          |
| APIGEE_AUTH_SERVERUSERNAME | The IDP username for requesting tokens           | edgecli                           |
| APIGEE_AUTH_SERVERPASSWORD | The IDP password for requesting tokens           | edgeclisecret                     |
