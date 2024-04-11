# Apigee Traceability Agent

The Traceability agent finds logs from consumed Apigee proxies and sends the traffic data to Amplify Central

## Build and run

The following make targets are available

| Target          | Description                                                    | Output(s)                        |
| --------------- | -------------------------------------------------------------- | -------------------------------- |
| dep             | downloads all dependencies needed to build the discovery agent | /vendor                          |
| test            | runs go test against all test files int he repo                | test results                     |
| update-sdk      | pulls the latest changes to main on the SDK repo               |                                  |
| build           | builds the binary traceability agent                           | bin/apigee_traceability_agent    |
| apigee-generate | generates the models for the Apigee APIs                       | pkg/apigee/models                |
| docker-build    | builds the traceability agent in a docker container            | apigee-traceability-agent:latest |

### Build (Docker)

```shell
make docker-build
```

### Run (Docker)

```shell
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

| Environment Variable       | Description                                                                                                              | Default (if applicable)           |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------ | --------------------------------- |
| APIGEE_URL                 | The base Apigee URL for this agent to connect to                                                                         | https://api.enterprise.apigee.com |
| APIGEE_APIVERSION          | The version of the API for the agent to use                                                                              | v1                                |
| APIGEE_DATAURL             | The base Apigee Data API URL for this agent to connect to                                                                | https://apigee.com/dapi/api       |
| APIGEE_ORGANIZATION        | The Apigee organization name                                                                                             |                                   |
| APIGEE_DEVELOPERID         | The Apigee developer, email, that will own all apps                                                                      |                                   |
| APIGEE_DISCOVERYMODE       | The mode in which the discovery agent operates, determines how stats are gathered, proxies (proxy) or products (product) | proxy                             |
| APIGEE_INTERVAL_STATS      | The polling interval checking for API Proxy changes, only in proxy mode                                                  | 5m (5 minutes), >=1m, <=15m       |
| APIGEE_AUTH_USERNAME       | The Apigee account username/email address                                                                                |                                   |
| APIGEE_AUTH_PASSWORD       | The Apigee account password                                                                                              |                                   |
| APIGEE_AUTH_USEBASICAUTH   | Set this to true to have the Apigee api client use HTTP Basic Authentication                                             | false                             |
| APIGEE_AUTH_URL            | The IDP URL                                                                                                              | https://login.apigee.com          |
| APIGEE_AUTH_SERVERUSERNAME | The IDP username for requesting tokens                                                                                   | edgecli                           |
| APIGEE_AUTH_SERVERPASSWORD | The IDP password for requesting tokens                                                                                   | edgeclisecret                     |
| APIGEE_FILTERED_APIS       | List that should contain apis for which metrics are wanted. Leave empty to use all the discovered apis instead           |                                   |
| APIGEE_FILTER_APIS         | Set to true if api metrics filtering is wanted                                                                           | false                             |

