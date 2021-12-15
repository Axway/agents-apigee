module github.com/Axway/agents-apigee/discovery

go 1.16

// replace github.com/Axway/agent-sdk => /home/ubuntu/go/src/github.com/Axway/agent-sdk
replace github.com/Axway/agents-apigee/client => ../client

require (
	github.com/Axway/agent-sdk v1.1.13-0.20211214195614-7f46c1019c8a
	github.com/Axway/agents-apigee/client v0.0.0-00010101000000-000000000000
	github.com/tidwall/gjson v1.12.1
)

replace (
	github.com/Shopify/sarama => github.com/elastic/sarama v0.0.0-20191122160421-355d120d0970
	github.com/dop251/goja => github.com/andrewkroh/goja v0.0.0-20190128172624-dd2ac4456e20
	github.com/fsnotify/fsevents => github.com/fsnotify/fsevents v0.1.1
)
