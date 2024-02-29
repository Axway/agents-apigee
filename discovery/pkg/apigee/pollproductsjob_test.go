package apigee

import (
	"fmt"
	"testing"
	"time"

	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/client/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_pollProductsJob(t *testing.T) {
	tests := []struct {
		name           string
		productName    string
		allProductErr  bool
		getProductErr  bool
		specNotFound   bool
		filterFailed   bool
		specNotInCache bool
		apiPublished   bool
	}{
		{
			name:         "api already published create update",
			apiPublished: true,
		},
		{
			name:        "api published with display name match",
			productName: "priv-PushNotif",
		},
		{
			name:        "api published with case insensitive name match",
			productName: "cell",
		},
		{
			name: "api published with name match",
		},
		{
			name:           "do not publish when spec was not in the cache",
			specNotInCache: true,
		},
		{
			name:         "do not publish when should publish check fails",
			filterFailed: true,
		},
		{
			name:          "should stop when getting product details fails",
			getProductErr: true,
		},
		{
			name:          "should stop when getting all products fails",
			allProductErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := mockProductClient{
				t:             t,
				productName:   tc.productName,
				allProductErr: tc.allProductErr,
				getProductErr: tc.getProductErr,
				specNotFound:  tc.specNotFound,
			}

			cache := mockProductCache{
				specNotInCache: tc.specNotInCache,
			}

			readyFunc := func() bool {
				return true
			}

			filterFunc := func(map[string]string) bool {
				return !tc.filterFailed
			}

			productJob := newPollProductsJob(client, cache, readyFunc, 10, filterFunc)
			assert.False(t, productJob.FirstRunComplete())

			productJob.isPublishedFunc = func(id string) bool {
				return tc.apiPublished
			}

			publishCalled := false
			// receive the publish call and validate what was published
			productJob.publishFunc = func(sb apic.ServiceBody) error {
				publishCalled = true
				return nil
			}

			err := productJob.Execute()
			if tc.allProductErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			// error getting all proxies should not flip first run
			if tc.allProductErr || tc.getProductErr || tc.filterFailed || tc.specNotInCache {
				assert.False(t, publishCalled)
			} else {
				assert.True(t, publishCalled)
			}
		})
	}
}

type mockProductClient struct {
	t             *testing.T
	cfg           *config.ApigeeConfig
	productName   string
	allProductErr bool
	getProductErr bool
	specNotFound  bool
}

func (m mockProductClient) GetConfig() *config.ApigeeConfig {
	return m.cfg
}

func (m mockProductClient) GetProducts() (products apigee.Products, err error) {
	productName := m.productName
	if productName == "" {
		productName = "RTE"
	}

	products = []string{productName}
	if m.allProductErr {
		products = nil
		err = fmt.Errorf("error get all products")
	}
	return
}

func (m mockProductClient) GetProduct(productName string) (*models.ApiProduct, error) {
	products := map[string]*models.ApiProduct{
		"RTE": {ApiResources: []string{},
			ApprovalType: "auto",
			Attributes: []models.Attribute{
				{
					Name:  "access",
					Value: "public",
				},
			},
			CreatedAt:      1665416157626,
			CreatedBy:      "cicd_technical_user@engie.com",
			Description:    "Generated Product",
			DisplayName:    "RTE",
			Environments:   []string{"acc", "int", "itt", "ppd"},
			LastModifiedAt: 1665758109625,
			LastModifiedBy: "cicd_technical_user@engie.com",
			Name:           "RTE",
			Proxies:        []string{"public-apiset-protected-ecowatt-v10"},
			Quota:          "10000",
			QuotaInterval:  "1",
			QuotaTimeUnit:  "minute",
			Scopes:         []string{"apihour:read", "apihour:write"},
		},
		"cell": {ApiResources: []string{"/"},
			ApprovalType: "auto",
			Attributes: []models.Attribute{
				{
					Name:  "access",
					Value: "public",
				},
			},
			CreatedAt:      1632752367332,
			CreatedBy:      "cicd_technical_user@engie.com",
			Description:    "Generated Product",
			DisplayName:    "Cell",
			Environments:   []string{"acc", "int", "itt", "ppd"},
			LastModifiedAt: 1665758109625,
			LastModifiedBy: "cicd_technical_user@engie.com",
			Name:           "cell",
			Proxies: []string{
				"public-cel-portefeuilles-contrats-v01",
				"public-cel-portefeuilles-contrats-v10",
				"public-cel-protected-adlperformance-v01",
				"public-cel-protected-pilotage-v01",
				"public-cel-protected-v01",
			},
			Quota:         "10000",
			QuotaInterval: "1",
			QuotaTimeUnit: "minute",
			Scopes:        []string{"apihour:read", "apihour:write"},
		},
		"priv-PushNotif": {ApiResources: []string{"/"},
			ApprovalType: "auto",
			Attributes: []models.Attribute{
				{
					Name:  "access",
					Value: "public",
				},
			},
			CreatedAt:      1632752359124,
			CreatedBy:      "cicd_technical_user@engie.com",
			Description:    "Generated Product",
			DisplayName:    "Private-PushNotif",
			Environments:   []string{"acc", "int", "itt", "ppd"},
			LastModifiedAt: 1665758129808,
			LastModifiedBy: "cicd_technical_user@engie.com",
			Name:           "priv-PushNotif",
			Proxies:        []string{"private-pushnotif-protected-airship-v10"},
			Quota:          "10000",
			QuotaInterval:  "1",
			QuotaTimeUnit:  "minute",
			Scopes:         []string{"apihour:read", "apihour:write"},
		},
	}
	if m.getProductErr {
		return nil, fmt.Errorf("error get product")
	}
	return products[productName], nil
}

func (m mockProductClient) GetSpecFile(path string) ([]byte, error) {
	assert.Equal(m.t, specPath, path)
	return []byte("spec"), nil
}

func (m mockProductClient) IsReady() bool { return false }

type mockProductCache struct {
	specNotInCache bool
}

func (m mockProductCache) GetSpecWithName(name string) (*specCacheItem, error) {
	if m.specNotInCache {
		return nil, fmt.Errorf("spec not in cache")
	}
	return &specCacheItem{
		ID:          "id",
		Name:        "name",
		ContentPath: "/path/to/spec",
		ModDate:     time.Now(),
	}, nil
}

func (m mockProductCache) AddPublishedServiceToCache(cacheKey string, serviceBody *apic.ServiceBody) {
}

func (m mockProductCache) AddProductToCache(name string, modDate time.Time, specHash string) {
}

func (m mockProductCache) HasProductChanged(name string, modDate time.Time, specHash string) bool {
	return true
}

func (m mockProductCache) GetProductWithName(name string) (*productCacheItem, error) {
	return nil, nil
}
