package apigee

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/cache"
)

type agentCache struct {
	cache              cache.Cache
	specEndpointToKeys map[string][]specCacheItem
	mutex              *sync.Mutex
}

type specCacheItem struct {
	ID          string
	Name        string
	ContentPath string
	ModDate     time.Time
}

func newAgentCache() *agentCache {
	return &agentCache{
		cache:              cache.New(),
		specEndpointToKeys: make(map[string][]specCacheItem),
		mutex:              &sync.Mutex{},
	}
}

func specPrimaryKey(name string) string {
	return fmt.Sprintf("spec-%s", name)
}

func (a *agentCache) AddSpecToCache(id, path, name string, modDate time.Time, endpoints ...string) {
	item := specCacheItem{
		ID:          id,
		Name:        strings.ToLower(name),
		ContentPath: path,
		ModDate:     modDate,
	}

	a.cache.SetWithSecondaryKey(specPrimaryKey(name), path, item)
	a.cache.SetSecondaryKey(specPrimaryKey(name), strings.ToLower(name))
	a.cache.SetSecondaryKey(specPrimaryKey(name), id)
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for _, ep := range endpoints {
		if _, found := a.specEndpointToKeys[ep]; !found {
			a.specEndpointToKeys[ep] = []specCacheItem{}
		}
		a.specEndpointToKeys[ep] = append(a.specEndpointToKeys[ep], item)
	}
}

func (a *agentCache) HasSpecChanged(name string, modDate time.Time) bool {
	data, err := a.cache.GetBySecondaryKey(name)
	if err != nil || data == nil {
		// spec not in cache
		return true
	}

	specItem := data.(specCacheItem)
	return modDate.After(specItem.ModDate)
}

func (a *agentCache) GetSpecWithPath(path string) (*specCacheItem, error) {
	data, err := a.cache.GetBySecondaryKey(path)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("spec path name %s not found in cache", path)
	}

	specItem := data.(specCacheItem)
	return &specItem, nil
}

func (a *agentCache) GetSpecWithName(name string) (*specCacheItem, error) {
	data, err := a.cache.GetBySecondaryKey(strings.ToLower(name))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("spec with name %s not found in cache", name)
	}

	specItem := data.(specCacheItem)
	return &specItem, nil
}

// GetSpecPathWithEndpoint - returns the lat modified spec found with this endpoint
func (a *agentCache) GetSpecPathWithEndpoint(endpoint string) (string, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	items, found := a.specEndpointToKeys[endpoint]
	if !found {
		return "", fmt.Errorf("no spec found for endpoint: %s", endpoint)
	}

	latest := specCacheItem{}
	for _, item := range items {
		if item.ModDate.After(latest.ModDate) {
			latest = item
		}
	}

	return latest.ContentPath, nil
}

func productPrimaryKey(name string) string {
	return fmt.Sprintf("product-%s", name)
}

func (a *agentCache) AddProductToCache(name string, modDate time.Time, specModDate time.Time) {
	item := productCacheItem{
		Name:        strings.ToLower(name),
		ModDate:     modDate,
		SpecModDate: specModDate,
	}

	a.cache.Set(productPrimaryKey(name), item)
}

func (a *agentCache) HasProductChanged(name string, modDate time.Time, specModDate time.Time) bool {
	data, err := a.cache.Get(productPrimaryKey(name))
	if err != nil || data == nil {
		// spec not in cache
		return true
	}

	productItem := data.(productCacheItem)
	return (modDate.After(productItem.ModDate) || specModDate.After(productItem.SpecModDate))
}

func (a *agentCache) GetProductWithName(name string) (*productCacheItem, error) {
	data, err := a.cache.Get(productPrimaryKey(name))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("product with name %s not found in cache", name)
	}

	productItem := data.(productCacheItem)
	return &productItem, nil
}

func (a *agentCache) AddPublishedServiceToCache(cacheKey string, serviceBody *apic.ServiceBody) {
	a.cache.Set(cacheKey, serviceBody)
}

func (a *agentCache) GetPublishedProxy(cacheKey string) (*apic.ServiceBody, error) {
	item, err := a.cache.Get(cacheKey)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("published proxy with key %s not found in cache", cacheKey)
	}

	sb := item.(*apic.ServiceBody)
	return sb, nil
}
