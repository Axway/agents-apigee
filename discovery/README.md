# Apigee Discovery Agent

The Discovery agent finds deployed API Proxies in Apigee then sends them to API Central

## Build and run

The following make targets are available

| Target     | Description                                                    | Output(s)                    |
|------------|----------------------------------------------------------------|------------------------------|
| lint       | runs go lint against all source files                          | linter results               |
| dep        | downloads all dependencies needed to build the discovery agent | /vendor                      |
| test       | runs go test against all test files int he repo                | test results                 |
| update-sdk | pulls the latest changes to main on the SDK repo               |                              |
| build      | builds the binary discovery agent                              | bn/apigee_tracebilityy_agent |