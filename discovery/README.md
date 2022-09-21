# Apigee Discovery Agent

The Discovery agent finds deployed API Proxies in Apigee then sends them to API Central

## Build and run

The following make targets are available

| Target          | Description                                                    | Output(s)                     |
| --------------- | -------------------------------------------------------------- | ----------------------------- |
| dep             | downloads all dependencies needed to build the discovery agent | /vendor                       |
| test            | runs go test against all test files int he repo                | test results                  |
| update-sdk      | pulls the latest changes to main on the SDK repo               |                               |
| build           | builds the binary discovery agent                              | bin/apigee_discovery_agent    |
| apigee-generate | generates the models for the Apigee APIs                       | pkg/apigee/models             |
| docker-build    | builds the discovery agent in a docker container               | apigee-discovery-agent:latest |

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

* Find all specs
* Find all Deployed API Proxies
  * Find the Spec
    * Proxy Revision has spec set, use it
    * Proxy Revision has association.json resource file, get path
      * Using path check to see if it is in the specs that were found by agent, use it
    * Using deployed URL path check for specs for match, use it
  * Check proxy for Key or Oauth policy for authentication
  * Create API Service
    * If spec was found use it in revision
    * If spec was not found create as unstructured  
    * Attach appropriate Credential Request Definition based on policy in proxy

## Discovery agent provisioning process

* Managed Application
  * Creates a new App on Apigee under the configured developer
* Access Request
  * Creates a new Product, or uses existing, based off the APIGEE-Proxy and Central Plan combination
  * Associates the new Product to any existing Credentials on the Application
* Credential
  * Creates a new Credential on the App and associates all Access Requests products to it

## Discovery agent variables

| Environment Variable  | Description                                               | Default (if applicable)           |
| --------------------- | --------------------------------------------------------- | --------------------------------- |
| APIGEE_ORGANIZATION   | The Apigee organization name                              |                                   |
| APIGEE_AUTH_USERNAME  | The Apigee account username/email address                 |                                   |
| APIGEE_AUTH_PASSWORD  | The Apigee account password                               |                                   |
| APIGEE_DEVELOPERID    | The Apigee developer, email, that will own all apps               |                                   |
| APIGEE_URL            | The base Apigee URL for this agent to connect to          | https://api.enterprise.apigee.com |
| APIGEE_APIVERSION     | The version of the API for the agent to use               | v1                                |
| APIGEE_DATAURL        | The base Apigee Data API URL for this agent to connect to | https://apigee.com/dapi/api       |
| APIGEE_INTERVAL_PROXY | The polling interval checking for API Proxy changes       | 30s (30 seconds)                  |
| APIGEE_INTERVAL_SPEC  | The polling interval for checking for new Specs           | 30m (30 minute)                   |
