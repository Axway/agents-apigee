# Prerequisites

* You need an Axway Platform user account that is assigned the AMPLIFY Central admin role
* Your Apigee Gateway/Manager should be up and running and have APIs to be discovered and exposed in AMPLIFY Central
* [Loggly](https://www.loggly.com/) account

Let’s get started!

## Prepare AMPLIFY Central Environments

In this section we'll:

* [Create an environment in Central](#create-an-environment-in-central)
* [Create a service account](#create-a-service-account)

### Create an environment in Central

* Log into [Amplify Central](https://apicentral.axway.com)
* Navigate to "Topology" then "Environments"
* Click "+ Environment"
  * Select a name
  * Click "Save"
* To enable the viewing of the agent status in Amplify see [Visualize the agent status](https://docs.axway.com/bundle/amplify-central/page/docs/connect_manage_environ/environment_agent_resources/index.html#add-your-agent-resources-to-the-environment)

### Create a service account

* Create a public and private key pair locally using the openssl command

```sh
openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits: 2048
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

* Log into the [Amplify Platform](https://platform.axway.com)
* Navigate to "Organization" then "Service Accounts"
* Click "+ Service Account"
  * Select a name
  * Optionally add a description
  * Select "Client Certificate"
  * Select "Provide public key"
  * Select or paste the contents of the public_key.pem file
  * Select "Central admin"
  * Click "Save"
* Note the Client ID value, this and the key files will be needed for the agents

## Prepare Apigee

* Create an Apigee account
* Note the username and password used as the agents will need this to run

## Prepare Loggly

* Create a [Loggly](https://www.loggly.com/) account
* Note the following values:
  * Subdomain
    * In Loggly select "Logs", on the left, then click on your profile icon on the upper right. It will list current subdomains
    * In the browser URL, "subdomain.loggly.com"
  * Customer Token
    * In Loggly select "Logs", on the left, then "Source Setup"
    * At the top select "Customer Tokens"
  * API Token
    * In Loggly select "Settings", on the left, "Log Settings". then "API Tokens"
    * Generate one if necessary

### Loggly integration with Apigee

In this section we'll describe the process the Discovery agent takes to setup Loggly in Apigee for the Traceability agent to capture traffic

On initial startup of the Discovery agent a new Apigee shared flow (amplify-central-logging) is created to capture the logs.

The shared flow has the following components.

* JavaScript execution step
  * Executed first in the flow to aggregate all request and response headers into a single variable which is saved on the flow
  * The following is the code executed

  ```javascript
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

* Message logging policy step
  * Executed to send transactional information to Loggly for the Traceability Agent to consume and push to Central
  * The following is the message logging policy data

  ```xml
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

* After the shared flow is created it is automatically deployed to each environment and added as a flow hook to execute Post Proxy

## Setup agent Environment Variables

The following environment variables file should be created for executing both of the agents

```ini
CENTRAL_ORGANIZATIONID=<Amplify Central Organization ID>
CENTRAL_TEAM=<Amplify Central Team Name>
CENTRAL_ENVIRONMENT=<Amplify Central Environment Name>   # created in Prepare AMPLIFY Central Environments step

CENTRAL_AUTH_CLIENTID=<Amplify Central Service Account>  # created in Prepare AMPLIFY Central Environments step
CENTRAL_AUTH_PRIVATEKEY=/keys/private_key.pem            # path to the key file created with openssl
CENTRAL_AUTH_PUBLICKEY=/keys/public_key.pem              # path to the key file created with openssl

APIGEE_ORGANIZATION=<Apigee Organization>                # created in Prepare Apigee step
APIGEE_AUTH_USERNAME=<Apigee Username>                   # created in Prepare Apigee step
APIGEE_AUTH_PASSWORD=<Apigee Password>                   # created in Prepare Apigee step

# Traceability agent only
LOGGLY_SUBDOMAIN=<Loggly Subdomain>               # created in Prepare Loggly step
LOGGLY_CUSTOMERTOKEN=<Loggly Customer Token>      # created in Prepare Loggly step
LOGGLY_APITOKEN=<Loggly API Token>                # created in Prepare Loggly step
LOGGLY_HOST=logs-01.loggly.com
LOGGLY_PORT=514

LOG_LEVEL=info
LOG_OUTPUT=stdout
```

## Discovery Agent

Reference: [Discovery Agent](/discovery/README.md)

## Traceability Agent

Reference: [Traceability Agent](/traceability/README.md)
