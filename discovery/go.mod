module github.com/Axway/agents-apigee/discovery

go 1.16

// replace github.com/Axway/agent-sdk => /home/ubuntu/go/src/github.com/Axway/agent-sdk

require (
	github.com/Axway/agent-sdk v1.1.27
	github.com/Axway/agents-apigee/client v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.5
)

replace (
	github.com/Axway/agents-apigee/client => ../client
	github.com/Shopify/sarama => github.com/elastic/sarama v1.19.1-0.20210823122811-11c3ef800752
	github.com/getkin/kin-openapi => github.com/getkin/kin-openapi v0.67.0
)
