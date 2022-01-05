# Apigee Discovery Agent

The Discovery agent finds deployed API Proxies in Apigee then sends them to API Central

![Discovery Agent Process](/resources/discovery_agent_apigee.JPG)

## Build and run

The following make targets are available

| Target          | Description                                                    | Output(s)                     |
|-----------------|----------------------------------------------------------------|-------------------------------|
| lint            | runs go lint against all source files                          | linter results                |
| dep             | downloads all dependencies needed to build the discovery agent | /vendor                       |
| test            | runs go test against all test files int he repo                | test results                  |
| update-sdk      | pulls the latest changes to main on the SDK repo               |                               |
| build           | builds the binary discovery agent                              | bin/apigee_discovery_agent    |
| apigee-generate | generates the models for the Apigee APIs                       | pkg/apigee/models             |
| build           | builds the traceability agent in a docker container            | apigee-discovery-agent:latest |

### Build (Docker)

```
make docker-build
```

### Run (Docker)

```
docker run --env-file env_vars -v `pwd`/keys:/keys apigee-discovery-agent:latest
```

### Build (Windows)

* Build the agent using the following command

```shell
go build -tags static_all \
    -ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=$${time}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=$${version}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=$${commit_id}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=APIGEEDiscoveryAgent'" \
    -a -o ./bin/apigee_discovery_agent.exe ./main.go
```

### Run (Windows)

* After a successful build, you should see the executable under the bin folder.   And you can execute it using the following command

```shell
./apigee_discovery_agent.exe --envFile env_vars
```

## Discovery agent proxy discovery

* Find all API portals within Apigee
* Within each portal find all published APIs
  * Using the published API get attributes from the attached product for filtering purposes
  * Get the Spec attached to the published API
* Create the Amplify Central API

## Discovery agent variables

| Environment Variable    | Description                                                  | Default (if applicable) |
|-------------------------|--------------------------------------------------------------|-------------------------|
| APIGEE_ORGANIZATION     | The Apigee organization name                                 |                         |
| APIGEE_AUTH_USERNAME    | The Apigee account username/email address                    |                         |
| APIGEE_AUTH_PASSWORD    | The Apigee account password                                  |                         |
| APIGEE_FILTER           | The tag filter to use against an Apigee product's attributes |                         |
| APIGEE_INTERVAL_PRODUCT | The time between getting a products attributes from Apigee   | 5m (5 minutes)          |
| APIGEE_INTERVAL_PORTAL  | The polling interval for new portals on Apigee               | 1m (1 minute)           |
| APIGEE_INTERVAL_API     | The polling interval for APIs in an Apigee portal            | 30s (30 seconds)        |
