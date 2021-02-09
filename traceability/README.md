# Apigee Traceability Agent

The Traceability agent finds deployed API Proxies in Apigee then sends them to API Central

## Build and run

The following make targets are available

| Target     | Description                                                    | Output(s)                    |
|------------|----------------------------------------------------------------|------------------------------|
| lint       | runs go lint against all source files                          | linter results               |
| dep        | downloads all dependencies needed to build the discovery agent | /vendor                      |
| test       | runs go test against all test files int he repo                | test results                 |
| update-sdk | pulls the latest changes to main on the SDK repo               |                              |
| build      | builds the binary traceability agent                           | bn/apigee_traceability_agent |

## Traceability agent variables

| Environment Variable        | Description                               | Default (if applicable) |
|-----------------------------|-------------------------------------------|-------------------------|
| APIGEE_ORGANIZATION         | The Apigee organization name              |                         |
| APIGEE_AUTH_USERNAME        | The Apigee account username/email address |                         |
| APIGEE_AUTH_PASSWORD        | The Apigee account password               |                         |
| APIGEE_LOGGLY_ORGANIZATION  | The Loggly organization name              |                         |
| APIGEE_LOGGLY_CUSTOMERTOKEN | The Loggly customer token                 |                         |
| APIGEE_LOGGLY_APITOKEN      | The Loggly API token                      |                         |
| APIGEE_LOGGLY_HOST          | The Loggly host address                   | logs-01.loggly.com      |
| APIGEE_LOGGLY_PORT          | The Loggly host port                      | 514                     |