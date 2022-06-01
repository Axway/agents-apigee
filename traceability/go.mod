module github.com/Axway/agents-apigee/traceability

go 1.16

// replace github.com/Axway/agent-sdk => /home/ubuntu/go/src/github.com/Axway/agent-sdk

replace github.com/Axway/agents-apigee/client => ../client

require (
	github.com/Axway/agent-sdk v1.1.25
	github.com/Axway/agents-apigee/client v0.0.0-00010101000000-000000000000
	github.com/Shopify/sarama v1.26.4 // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/elastic/beats/v7 v7.17.2
	github.com/stretchr/testify v1.7.1
	github.com/urso/magetools v0.0.0-20200106130147-61080ed7b22b // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/Shopify/sarama => github.com/elastic/sarama v1.19.1-0.20210823122811-11c3ef800752
	github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20190717161051-705d9623b7c1+incompatible
	github.com/dop251/goja => github.com/andrewkroh/goja v0.0.0-20190128172624-dd2ac4456e20
	github.com/getkin/kin-openapi => github.com/getkin/kin-openapi v0.67.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.1
	k8s.io/client-go => k8s.io/client-go v0.21.1
)
