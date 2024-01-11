package apigee

import (
	"testing"

	"github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/client/pkg/config"
	"github.com/stretchr/testify/assert"
)

func createTestClient(t *testing.T, mockClient *api.MockHTTPClient) *ApigeeClient {
	cfg := &config.ApigeeConfig{
		URL:          "http://test.com",
		APIVersion:   "v1",
		Organization: "org",
		DeveloperID:  "test@dev.id",
		DataURL:      "http://data.com",
		Auth: &config.AuthConfig{
			BasicAuth: true,
			Username:  "user",
			Password:  "pass",
		},
	}

	c, err := NewClient(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, c.GetConfig(), cfg)
	c.apiClient = mockClient

	return c
}

func TestNewClient(t *testing.T) {
	c := createTestClient(t, nil)

	assert.Equal(t, c.GetDeveloperID(), "test@dev.id")
	assert.True(t, c.IsReady())
}
