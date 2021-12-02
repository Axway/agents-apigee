module github.com/Axway/agents-apigee/discovery

go 1.16

require (
	github.com/Axway/agent-sdk v1.1.11-0.20211202200236-6b12e662e18a
	github.com/getkin/kin-openapi v0.67.0
	github.com/markbates/pkger v0.17.1
	github.com/oriser/regroup v0.0.0-20201024192559-010c434ff8f3
	github.com/tidwall/gjson v1.9.4
)

replace (
	github.com/Shopify/sarama => github.com/elastic/sarama v0.0.0-20191122160421-355d120d0970
	github.com/dop251/goja => github.com/andrewkroh/goja v0.0.0-20190128172624-dd2ac4456e20
	github.com/fsnotify/fsevents => github.com/fsnotify/fsevents v0.1.1
)
