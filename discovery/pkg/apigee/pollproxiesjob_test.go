package apigee

import (
	"fmt"
	"testing"

	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/apic/provisioning"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/stretchr/testify/assert"
)

const (
	proxyName    = "string"
	envName      = "prod"
	revName      = "1"
	specPath     = "/path/to/spec"
	fullSpecPath = "http://host.com/path/to/spec"
	apiKeyName   = "apiKeyPolicy"
	oauthName    = "oauthPolicy"
)

func Test_pollProxiesJob(t *testing.T) {
	tests := []struct {
		name             string
		allProxyErr      bool
		getDeploymentErr bool
		getRevisionErr   bool
		specFound        bool
		revSpec          bool
		fullSpec         bool
		specInResource   bool
		hasAPIKey        bool
		hasOauth         bool
	}{
		{
			name:           "should create proxy when spec in revision resource file",
			specFound:      true,
			specInResource: true,
			hasOauth:       true,
			hasAPIKey:      true,
		},
		{
			name:      "should create proxy when spec url on revision",
			specFound: true,
			revSpec:   true,
		},
		{
			name:      "should create proxy when spec has full url on revision",
			specFound: true,
			revSpec:   true,
			fullSpec:  true,
		},
		{
			name:     "should create proxy when no spec found but has oauth policy",
			hasOauth: true,
		},
		{
			name: "should create proxy when no spec found",
		},
		{
			name: "should stop when no spec found but has api key policy",
		},
		{
			name:           "should stop when getting proxy revision fails",
			getRevisionErr: true,
		},
		{
			name:             "should stop when getting proxy deployment fails",
			getDeploymentErr: true,
		},
		{
			name:        "should stop when getting all proxies fails",
			allProxyErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := mockProxyClient{
				t:                t,
				allProxyErr:      tc.allProxyErr,
				getDeploymentErr: tc.getDeploymentErr,
				getRevisionErr:   tc.getRevisionErr,
				revSpec:          tc.revSpec,
				fullSpec:         tc.fullSpec,
				specInResource:   tc.specInResource,
				hasAPIKey:        tc.hasAPIKey,
				hasOauth:         tc.hasOauth,
			}
			proxyJob := newPollProxiesJob(client, mockProxyCache{}, func() bool { return true }, 10)
			assert.False(t, proxyJob.FirstRunComplete())

			// receive the publish call and validate what was published
			proxyJob.publishFunc = func(sb apic.ServiceBody) error {
				assert.Equal(t, proxyName, sb.APIName)
				assert.Equal(t, proxyName, sb.RestAPIID)
				assert.Equal(t, envName, sb.Stage)
				assert.Equal(t, revName, sb.Version)
				assert.Equal(t, "A Proxy", sb.NameToPush)
				assert.Equal(t, "A Proxy Description", sb.Description)
				crds := []string{}
				if tc.hasAPIKey {
					crds = append(crds, provisioning.APIKeyCRD)
				}
				if tc.hasOauth {
					crds = append(crds, provisioning.OAuthSecretCRD)
				}
				assert.Equal(t, crds, sb.GetCredentialRequestDefinitions())

				if tc.specFound {
					assert.NotEmpty(t, sb.SpecDefinition)
				} else {
					assert.Empty(t, sb.SpecDefinition)
				}
				return nil
			}

			proxyJob.Execute()

			// error getting all proxies should not flip first run
			assert.NotEqual(t, tc.allProxyErr, proxyJob.FirstRunComplete())
		})
	}
}

type mockProxyClient struct {
	t                *testing.T
	allProxyErr      bool
	getDeploymentErr bool
	getRevisionErr   bool
	revSpec          bool
	fullSpec         bool
	specInResource   bool
	hasAPIKey        bool
	hasOauth         bool
}

func (m mockProxyClient) GetAllProxies() (proxies apigee.Proxies, err error) {
	proxies = []string{proxyName}
	if m.allProxyErr {
		proxies = nil
		err = fmt.Errorf("error")
	}
	return
}

func (m mockProxyClient) GetDeployments(apiName string) (deployment *models.DeploymentDetails, err error) {
	assert.Contains(m.t, proxyName, apiName)
	deployment = &models.DeploymentDetails{
		Environment: []models.DeploymentDetailsEnvironment{
			{
				Name: envName,
				Revision: []models.DeploymentDetailsRevision{
					{
						Name: revName,
					},
				},
			},
		},
	}
	if m.getDeploymentErr {
		deployment = nil
		err = fmt.Errorf("error")
	}
	return
}

func (m mockProxyClient) GetRevision(apiName, revision string) (rev *models.ApiProxyRevision, err error) {
	assert.Contains(m.t, proxyName, apiName)
	assert.Contains(m.t, revName, revision)
	rev = &models.ApiProxyRevision{
		Name:        proxyName,
		DisplayName: "A Proxy",
		Revision:    revName,
		Description: "A Proxy Description",
		Policies:    []string{},
	}
	if m.revSpec {
		rev.Spec = specPath
	}
	if m.fullSpec {
		rev.Spec = fullSpecPath
	}
	if m.specInResource {
		rev.ResourceFiles = models.ApiProxyRevisionResourceFiles{
			ResourceFile: []models.ApiProxyRevisionResourceFilesResourceFile{
				{
					Type: openapi,
					Name: association,
				},
			},
		}
	}
	if m.hasAPIKey {
		rev.Policies = append(rev.Policies, apiKeyName)
	}
	if m.hasOauth {
		rev.Policies = append(rev.Policies, oauthName)
	}
	if m.getRevisionErr {
		rev = nil
		err = fmt.Errorf("error")
	}
	return
}

func (m mockProxyClient) GetRevisionResourceFile(apiName, revision, resourceType, resourceName string) ([]byte, error) {
	assert.Contains(m.t, proxyName, apiName)
	assert.Contains(m.t, revName, revision)
	assert.Contains(m.t, openapi, resourceType)
	assert.Contains(m.t, association, resourceName)
	return []byte(fmt.Sprintf(`{
		"url": "%s"
	}`, specPath)), nil
}

func (m mockProxyClient) GetVirtualHost(envName, virtualHostName string) (*models.VirtualHost, error) {
	return nil, nil
}

func (m mockProxyClient) GetSpecFile(path string) ([]byte, error) {
	assert.Equal(m.t, specPath, path)
	return []byte("spec"), nil
}

func (m mockProxyClient) GetSpecFromURL(url string, options ...apigee.RequestOption) ([]byte, error) {
	assert.Equal(m.t, fullSpecPath, url)
	return []byte("spec"), nil
}

func (m mockProxyClient) GetRevisionPolicyByName(apiName, revision, policyName string) (policy *apigee.PolicyDetail, err error) {
	assert.Contains(m.t, proxyName, apiName)
	assert.Contains(m.t, revName, revision)
	if policyName == apiKeyName {
		policy = &apigee.PolicyDetail{
			PolicyType: apiKeyPolicy,
		}
	}
	if policyName == oauthName {
		policy = &apigee.PolicyDetail{
			PolicyType: oauthPolicy,
		}
	}
	return
}

func (m mockProxyClient) IsReady() bool { return false }

type mockProxyCache struct{}

func (m mockProxyCache) GetSpecWithPath(path string) (*specCacheItem, error) {
	return &specCacheItem{}, nil
}

func (m mockProxyCache) GetSpecPathWithEndpoint(endpoint string) (string, error) {
	return "", nil
}

func (m mockProxyCache) AddPublishedServiceToCache(cacheKey string, serviceBody *apic.ServiceBody) {}
