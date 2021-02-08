# Prerequisites
 * You need an Axway Platform user account that is assigned the AMPLIFY Central admin role
 * Your apigee Gateway/Manager should be up and running and have APIs to be discovered and exposed in AMPLIFY Central
 * GoLang 
 * [Loggly](https://www.loggly.com/)
 
Let’s get started!

## Prepare AMPLIFY Central Environments

In this section we'll:
 * Make sure we have access to the platform
 * Create a Platform team with user(s) for this activity (connecting Central to the Gateways)
 * Create a Central Environment for each Gateway we want to connect to
 * Create a service account for the agents to have programmatic, secure access to Central without having to use our login credentials

 - Create an Axway platform account
     - Create a [free trial account](https://platform.axway.com/) if you don't already have one
 - Create a Team
    - Once you are logged into the platform, you are in an organization. If you are a member of more than one organization, then you can switch organizations to the organization you want to work in
   - Create a Team called apigee-agents and add yourself (and others involved in this activity) to this team. We will use this team name later for configuring the agents
   - You must have Platform Administrator role for this activity
 - Create an environment in Central that corresponds to your Gateway (apigee).  We will create a unique environment for each Gateway, In case your connecting multiple gateways 
  - Create an environment called apigee. Note the name and the title. The name will be referenced later in the agent configuration
   - In the portal, go to Central -> Topology > Environments
   - Add new + Environment
 - Create a Service Account in Central so that the Agents can connect to the Gateway without using/exposing your client credentials. We can use the same service account for all Gateways/Agents
   - On your computer, generate a private and public key pair using these two commands
   - This key pair is used to setup service account and later to configure the agents
   - In the portal navigate to Central > Access> Service Accounts
   - Create service account using the public key created above (cat public_key.pem and copy the output)
   - Copy the client identified from Client ID field (DOSA_XXXXXX). You will use this in the following steps
 
## Prepare your APIGEE Gateway/Manager
 - Your apigee Gateway/Manager should be up and running and have APIs to be discovered and exposed in AMPLIFY Central
 - If you don’t have an apigee account yet.  You can create free account using the following [link]()
 - For sample API’s to import into your apigee account you can use the PetStore swagger using the following link 
 
## Agents Setup and Installation
In this section we'll setup the Discovery Agent and the Traceability Agent so that APIs and their traffic will be visible in Central and the Unified Catalog.

For the discovery and traceability agents to be configured.  You must create an amplify-central-logging shared flow, with steps:
- JavaScript script that aggregates all the headers and creates variables for the flow
```
 // Read request headers for both proxy and target flow
 var headerNames = context.getVariable('request.headers.names');
 var strHeaderNames = String(headerNames);
 var headerList = strHeaderNames.substring(1, strHeaderNames.length - 1).split(new RegExp(', ', 'g'));
 var reqHeaders = {};
 headerList.forEach(function(headerName) {
   reqHeaders[headerName] = context.getVariable('request.header.' + headerName);
 });
 // Read response headers for proxy flow
 headerNames = context.getVariable('response.headers.names');
 strHeaderNames = String(headerNames);
 headerList = strHeaderNames.substring(1, strHeaderNames.length - 1).split(new RegExp(', ', 'g'));
 var resHeaders = {};
 headerList.forEach(function(headerName) {
   resHeaders[headerName] = context.getVariable('response.header.' + headerName);
 });
 
 context.setVariable("apic.reqHeaders", JSON.stringify(JSON.stringify(reqHeaders)));
 context.setVariable("apic.resHeaders", JSON.stringify(JSON.stringify(resHeaders)));
```
- Message logging policy that takes those headers as well as other info to send to a centralized logging server (Loggly in our development).  Below is the code snippets for message loggin policy
```
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<MessageLogging async="false" continueOnError="false" enabled="true" name="amplify-central-logging">
    <DisplayName>amplify-central-logging</DisplayName>
    <Syslog>
        <Message>[<LogglyToken>@41058 tag="apic-logs"]{
                 "organization":"{organization.name}",
                 "environment": "{environment.name}",
                 "api": "{apiproxy.name}",
                 "revision": "{apiproxy.revision}",
                 "messageId": "{messageid}",
                 "verb": "{request.verb}",
                 "path": "{request.path}",
                 "queryString": "{request.querystring}",
                 "clientIP": "{client.ip}",
                 "clientHost": "{client.host}",
                 "clientStartTimeStamp": "{client.received.start.timestamp}",
                 "clientEndTimeStamp": "{system.timestamp}",
                 "bytesReceived": "{request.header.Content-Length}",
                 "bytesSent": "{response.header.Content-Length}",
                 "userAgent": "{request.header.User-Agent}",
                 "httpVersion": "{request.version}",
                 "isError": "{is.error}",
                 "statusCode": "{response.status.code}",
                 "errorStatusCode": "{error.status.code}",
                 "requestHost":"{request.header.Host}",
                 "responseHost":"{response.header.Host}",
                 "requestHeaders": {apic.reqHeaders},
                 "responseHeaders": {apic.resHeaders}
                 }</Message>
        <Host><<Loggly Host>></Host>
        <Port><<Loggly Port>></Port>
        <FormatMessage>true</FormatMessage>
    </Syslog>
</MessageLogging>
```
- Deployed the shared flow to all environments
- Add the shared flow as a Flow Hook, Post Proxy, on all environments
- Main discovery loop checking for new and changed proxies

## Discovery Agent
#### Steps to implement discovery agent using this stub
1. Locate the commented tag "CHANGE_HERE" for package import paths in all files and fix them to reference the path correctly.
2. Run "make dep" to resolve the dependencies. This should resolve all dependency packages and vendor them under ./vendor directory
3. Update Makefile to change the name of generated binary image from *apic_discovery_agent* to the desired name. Locate *apic_discovery_agent* string and replace with desired name
4. Update pkg/cmd/root.go to change the name and description of the agent. Locate *apic_discovery_agent* and *Sample Discovery Agent* and replace to desired values
5. Update pkg/config/config.go to define the gateway specific configuration
    - Locate *gateway-section* and replace with the name of your gateway. Same string in pkg/cmd/root.go and sample YAML config file
    - Define gateway specific config properties in *GatewayConfig* struct. Locate the struct variables *ConfigKey1* & struct *config_key_1* and add/replace desired config properties
    - Add config validation. Locate *ValidateCfg()* method and update the implementation to add validation specific to gateway specific config.
    - Update the config binding with command line flags in init(). Locate *gateway-section.config_key_1* and add replace desired config property bindings
    - Update the initialization of gateway specific by parsing the binded properties. Locate *ConfigKey1* & *gateway-section.config_key_1* and add/replace desired config properties
6. Update pkg/gateway/client.go to implement the logic to discover and fetch the details related of the APIs.
    - Locate *DiscoverAPIs()* method and implement the logic
    - Locate *buildServiceBody()* method and update the Set*() method according to the API definition from gateway
7. Run "make build" to build the agent
8. Rename *apic_discovery_agent.yaml* file to the desired agents name and setup the agent config in the file.
9. Execute the agent by running the binary file generated under *bin* directory. The YAML config must be in the current working directory 

Reference: [SDK Documentation - Building Discovery Agent](https://github.com/Axway/agent-sdk/blob/main/docs/discovery/index.md)

## Traceability Agent 
#### Steps to implement traceability agent using this stub
1. Locate the commented tag "CHANGE_HERE" for package import paths in all files and fix them to reference the path correctly.
2. Run "make dep" to resolve the dependencies. This should resolve all dependency packages and vendor them under ./vendor directory
3. Update Makefile to change the name of generated binary image from *apic_traceability_agent* to the desired name. Locate *apic_traceability_agent* string and replace with desired name
4. Update pkg/cmd/root.go to change the name and description of the agent. Locate *apic_traceability_agent* and *Sample Traceability Agent* and replace to desired values
5. Update pkg/config/config.go to define the gateway specific configuration
    - Locate *gateway-section* and replace with the name of your gateway. Same string in pkg/cmd/root.go and sample YAML config file
    - Define gateway specific config properties in *GatewayConfig* struct. Locate the struct variables *ConfigKey1* & struct *config_key_1* and add/replace desired config properties
    - Add config validation. Locate *ValidateCfg()* method and update the implementation to add validation specific to gateway specific config.
    - Update the config binding with command line flags in init(). Locate *gateway-section.config_key_1* and add replace desired config property bindings
    - Update the initialization of gateway specific by parsing the binded properties. Locate *ConfigKey1* & *gateway-section.config_key_1* and add/replace desired config properties
6. Locate pkg/gateway/definition.go to define the structure of the log entry the traceability agent will receive. See pkg/gateway/definition.go for sample definition.
7. Implement the mechanism to read the log entry. Optionally you can wrap the existing beat(for e.g. filebeat) in beater.New() to read events and they setup output event processor to process the events
8. Locate pkg/gateway/eventprocessor.go to perform processing on event to be published. The processing can be performed either on the received event by beat input or before the event is published by transport. See pkg/gateway/eventprocessor.go for example of both type of processing.
9. Locate pkg/gateway/eventmapper.go to map the log entry received by beat to event structure expected for AMPLIFY Central Observer.
10. Run "make build" to build the agent.
11. Rename *apic_traceability_agent.yml* file to the desired agents name and setup the agent config in the file.
12. Execute the agent by running the binary file generated under *bin* directory. The YAML config must be in the current working directory
13. To produce traffic update the ./logs/traffic.log file with a new entry. See ./logs/traffic.log for sample entries

Reference: [SDK Documentation - Building Traceability Agent](https://github.com/Axway/agent-sdk/blob/main/docs/traceability/index.md)
