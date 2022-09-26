package apigee

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

// isFullURL - returns true if the url arg is a fully qualified URL
func isFullURL(urlString string) bool {
	if parsed, err := url.ParseRequestURI(urlString); err == nil {
		return (parsed.Host != "" && parsed.Scheme != "")
	}
	return false
}

func urlsFromVirtualHost(virtualHost *models.VirtualHost, proxyName string) []string {
	urls := []string{}

	scheme := "http"
	port := virtualHost.Port
	if virtualHost.SSLInfo != nil {
		scheme = "https"
		if port == "443" {
			port = ""
		}
	}
	if scheme == "http" && port == "80" {
		port = ""
	}

	for _, host := range virtualHost.HostAliases {
		thisURL := fmt.Sprintf("%s://%s:%s", scheme, host, port)
		if port == "" {
			thisURL = fmt.Sprintf("%s://%s", scheme, host)
		}
		if virtualHost.BaseUrl != "/" {
			thisURL += virtualHost.BaseUrl
		}
		thisURL += "/" + strings.ToLower(proxyName)
		urls = append(urls, thisURL)
	}

	return urls
}

func createProxyCacheKey(id, envName string) string {
	return fmt.Sprintf("apiproxy-%s-%s", envName, id)
}

type ctxKeys string

const (
	loggerKey ctxKeys = "logger"
)

func (c ctxKeys) String() string {
	return string(c)
}

func addLoggerToContext(ctx context.Context, logger log.FieldLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func getLoggerFromContext(ctx context.Context) log.FieldLogger {
	return ctx.Value(loggerKey).(log.FieldLogger)
}

func getStringFromContext(ctx context.Context, key ctxKeys) string {
	return ctx.Value(key).(string)
}

func createEndpointsFromURLS(urls []string) []apic.EndpointDefinition {
	endpoints := []apic.EndpointDefinition{}

	for _, ep := range urls {
		u, err := url.Parse(ep)
		if err != nil {
			continue
		}
		port := int64(0)
		if p := u.Port(); p != "" {
			pt, err := strconv.ParseInt(p, 10, 32)
			if err == nil {
				port = pt
			}
		}
		endpoints = append(endpoints, apic.EndpointDefinition{
			Host:     u.Host,
			Port:     int32(port),
			BasePath: u.Path,
			Protocol: u.Scheme,
		})
	}
	return endpoints
}
