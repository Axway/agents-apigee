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