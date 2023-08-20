package cache_test

import (
	"testing"

	carbon "github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stretchr/testify/assert"
)

func TestCacheSetAndGet(t *testing.T) {
	c := cache.New("stackup-test", "/tmp")
	defer c.Cleanup(true)

	expires := carbon.Now().SubMinutes(5)
	entry := c.CreateEntry("test", "test", &expires, "", "", nil)

	c.Set("test", entry, -2)
	_, found := c.Get("test")
	assert.False(t, found)
	assert.True(t, c.IsExpired("test"))

	expires = carbon.Now().AddMinutes(5)
	entry2 := c.CreateEntry("test2", "test2", &expires, "", "", nil)

	c.Set("test2", entry2, 5)
	found2 := c.Has("test2")
	assert.True(t, found2)
}

func TestCacheRemove(t *testing.T) {
	c := cache.New("stackup-test", "/tmp")
	defer c.Cleanup(true)

	expires := carbon.Now().AddMinutes(5)
	entry := c.CreateEntry("test3", "test3", &expires, "", "", nil)

	c.Set("test3", entry, 5)
	assert.True(t, c.Has("test3"))

	c.Remove("test3")
	assert.False(t, c.Has("test3"))
}
