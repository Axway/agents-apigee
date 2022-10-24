package apigee

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_cacheSpecs(t *testing.T) {
	// create new agent cache
	c := newAgentCache()
	assert.NotNil(t, c)

	// add specs to cache
	c.AddSpecToCache("id1", "/path/id1", "name-id1", time.Now())
	c.AddSpecToCache("id2", "/path/id2", "name-id2", time.Now(), "http://id2/endpoint1", "http://id2/endpoint2")

	// get spec items

	// exists
	//path
	item, err := c.GetSpecWithPath("/path/id1")
	assert.NotNil(t, item)
	assert.Nil(t, err)
	//name
	item, err = c.GetSpecWithName("name-id2")
	assert.NotNil(t, item)
	assert.Nil(t, err)
	//endpoint
	path, err := c.GetSpecPathWithEndpoint("http://id2/endpoint1")
	assert.NotEmpty(t, path)
	assert.Nil(t, err)

	// not exists
	//path
	item, err = c.GetSpecWithPath("/path/id3")
	assert.Nil(t, item)
	assert.NotNil(t, err)
	//name
	item, err = c.GetSpecWithName("name-id3")
	assert.Nil(t, item)
	assert.NotNil(t, err)
	//endpoint
	path, err = c.GetSpecPathWithEndpoint("http://id2/endpoint3")
	assert.Empty(t, path)
	assert.NotNil(t, err)

	// has spec changed
	changed := c.HasSpecChanged("name-id1", time.Now().Add(time.Hour))
	assert.True(t, changed)
	changed = c.HasSpecChanged("name-id1", time.Now().Add(-1*time.Hour))
	assert.False(t, changed)
	changed = c.HasSpecChanged("name-id3", time.Now().Add(-1*time.Hour)) // doesn't exist is changed
	assert.True(t, changed)
}

func Test_cacheProducts(t *testing.T) {
	// create new agent cache
	c := newAgentCache()
	assert.NotNil(t, c)

	// add products to cache
	c.AddProductToCache("prod1", time.Now(), time.Now())

	// get product item
	//exists
	item, err := c.GetProductWithName("prod1")
	assert.NotNil(t, item)
	assert.Nil(t, err)
	//not exists
	item, err = c.GetProductWithName("prod2")
	assert.Nil(t, item)
	assert.NotNil(t, err)

	// has product changed
	//product change
	changed := c.HasProductChanged("prod1", time.Now().Add(time.Hour), time.Now().Add(-1*time.Hour))
	assert.True(t, changed)
	//spec change
	changed = c.HasProductChanged("prod1", time.Now().Add(-1*time.Hour), time.Now().Add(time.Hour))
	assert.True(t, changed)
	//no change
	changed = c.HasProductChanged("prod1", time.Now().Add(-1*time.Hour), time.Now().Add(-1*time.Hour))
	assert.False(t, changed)
	//spec change
	changed = c.HasProductChanged("prod2", time.Now().Add(-1*time.Hour), time.Now().Add(-1*time.Hour)) // no match returns changed
	assert.True(t, changed)
}
