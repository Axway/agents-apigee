package apigee

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

// isFullURL - returns true if the url arg is a fully qualified URL
func isFullURL(urlString string) bool {
	if _, err := url.ParseRequestURI(urlString); err != nil {
		return true
	}
	return false
}

func urlsFromVirtualHost(virtualHost *models.VirtualHost) []string {
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
