# Apigee Discovery Agent

The Discovery agent finds deployed API Proxies in Apigee then sends them to API Central

## Build and run

The following make targets are available

| Target          | Description                                                    | Output(s)                 |
|-----------------|----------------------------------------------------------------|---------------------------|
| lint            | runs go lint against all source files                          | linter results            |
| dep             | downloads all dependencies needed to build the discovery agent | /vendor                   |
| test            | runs go test against all test files int he repo                | test results              |
| update-sdk      | pulls the latest changes to main on the SDK repo               |                           |
| build           | builds the binary discovery agent                              | bn/apigee_discovery_agent |
| apigee-generate | generates the models for the Apigee APIs                       | pkg/apigee/models         |

## Discovery agent proxy discovery

* Find all deployed API proxies
* Download the proxy bundle
  * Find or generate a spec
    * Associated spec
      * Check for an association.json file in the apiproxy/resources/openapi path
      * Open the file to find the spec url path
      * Download the spec from Apigee
    * Generate spec
      * Read in all proxy endpoints
      * Create OAS3 spec with information about endpoints and policies
* Create the Amplify Central API

## Steps to run the discovery agent

* Build the agent using the following command
```
make build

```
If you don't have to Make configured.  You can use the go build command as follows for Windows
```
go build -tags static_all \
    -ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=$${time}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=$${version}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=$${commit_id}' \
        -X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=APIGEEDiscoveryAgent'" \
    -a -o ${WORKSPACE}/bin/apigee_discovery_agent.exe ${WORKSPACE}/main.go

```

* After a successful build, you should see the executable under the bin folder.   And you can execute it using the following command

```
./apigee_discovery_agent.exe --envFile env_vars
