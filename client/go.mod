module github.com/Axway/agents-apigee/client

go 1.16

// replace github.com/Axway/agent-sdk => /home/ubuntu/go/src/github.com/Axway/agent-sdk

require (
	github.com/Axway/agent-sdk v1.1.18-0.20220328224127-722da35d5f8a
	github.com/hashicorp/hcl v1.0.1-0.20180906183839-65a6292f0157 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	golang.org/x/sys v0.0.0-20220222172238-00053529121e // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/Shopify/sarama => github.com/elastic/sarama v0.0.0-20191122160421-355d120d0970
	github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20190717161051-705d9623b7c1+incompatible
	github.com/docker/go-plugins-helpers => github.com/elastic/go-plugins-helpers v0.0.0-20200207104224-bdf17607b79f
	github.com/dop251/goja => github.com/andrewkroh/goja v0.0.0-20190128172624-dd2ac4456e20
	github.com/fsnotify/fsevents => github.com/elastic/fsevents v0.0.0-20181029231046-e1d381a4d270
	github.com/fsnotify/fsnotify => github.com/adriansr/fsnotify v0.0.0-20180417234312-c9bbe1f46f1d
	github.com/google/gopacket => github.com/adriansr/gopacket v1.1.18-0.20200327165309-dd62abfa8a41
	github.com/insomniacslk/dhcp => github.com/elastic/dhcp v0.0.0-20200227161230-57ec251c7eb3 // indirect
	github.com/tonistiigi/fifo => github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c
	k8s.io/api => k8s.io/api v0.17.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.0
	k8s.io/client-go => k8s.io/client-go v0.17.0
)
