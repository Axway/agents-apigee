package apigee

import (
	"fmt"
	"strings"
	"time"

	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/cache"
)

type agentCache struct {
	cache              cache.Cache
	specEndpointToKeys map[string][]specCacheItem
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
	}
}

func (a *agentCache) AddSpecToCache(id, path, name string, modDate time.Time, endpoints ...string) {
	item := specCacheItem{
		ID:          id,
		Name:        strings.ToLower(name),
		ContentPath: path,
		ModDate:     modDate,
	}

	a.cache.SetWithSecondaryKey(name, path, item)
	a.cache.SetSecondaryKey(name, strings.ToLower(name))
	a.cache.SetSecondaryKey(name, id)
	for _, ep := range endpoints {
		if _, found := a.specEndpointToKeys[ep]; !found {
			a.specEndpointToKeys[ep] = []specCacheItem{}
		}
		a.specEndpointToKeys[ep] = append(a.specEndpointToKeys[ep], item)
	}
}

func (a *agentCache) HasSpecChanged(name string, modDate time.Time) bool {
	data, err := a.cache.GetBySecondaryKey(name)
	if err != nil {
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

	specItem := data.(specCacheItem)
	return &specItem, nil
}

func (a *agentCache) GetSpecWithName(name string) (*specCacheItem, error) {
	data, err := a.cache.GetBySecondaryKey(strings.ToLower(name))
	if err != nil {
		return nil, err
	}

	specItem := data.(specCacheItem)
	return &specItem, nil
}

// GetSpecPathWithEndpoint - returns the lat modified spec found with this endpoint
func (a *agentCache) GetSpecPathWithEndpoint(endpoint string) (string, error) {
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

func (a *agentCache) AddPublishedServiceToCache(cacheKey string, serviceBody *apic.ServiceBody) {
	a.cache.Set(cacheKey, serviceBody)
}

func (a *agentCache) GetPublishedProxy(cacheKey string) (*apic.ServiceBody, error) {
	item, err := a.cache.Get(cacheKey)
	if err != nil {
		return nil, err
	}

	sb := item.(*apic.ServiceBody)
	return sb, nil
}
