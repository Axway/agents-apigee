module github.com/Axway/agents-apigee/client

go 1.16

// replace github.com/Axway/agent-sdk => /home/ubuntu/go/src/github.com/Axway/agent-sdk

require (
	github.com/Axway/agent-sdk v1.1.22-0.20220513174935-b36ada3ea5e3
	github.com/hashicorp/hcl v1.0.1-0.20180906183839-65a6292f0157 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220222172238-00053529121e // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/Shopify/sarama => github.com/elastic/sarama v1.19.1-0.20210823122811-11c3ef800752
	github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20190717161051-705d9623b7c1+incompatible
	github.com/getkin/kin-openapi => github.com/getkin/kin-openapi v0.67.0
)
