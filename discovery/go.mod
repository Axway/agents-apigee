module github.com/Axway/agents-apigee/discovery

go 1.16

// replace github.com/Axway/agent-sdk => /home/ubuntu/go/src/github.com/Axway/agent-sdk

replace github.com/Axway/agents-apigee/client => ../client

require (
	github.com/Axway/agent-sdk v1.1.15-0.20220204124216-fb9fc6fe71b2
	github.com/Axway/agents-apigee/client v0.0.0-00010101000000-000000000000
)

replace (
	github.com/Shopify/sarama => github.com/elastic/sarama v0.0.0-20191122160421-355d120d0970
	github.com/dop251/goja => github.com/andrewkroh/goja v0.0.0-20190128172624-dd2ac4456e20
	github.com/fsnotify/fsevents => github.com/fsnotify/fsevents v0.1.1
)
