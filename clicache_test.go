package clicache

import (
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	// This is a temporary cleanup utility to ensure no cache files are left after the test
	defer func() {
		files, _ := filepath.Glob("/tmp/" + cachePrefix + "*.gob")
		for _, file := range files {
			os.Remove(file)
		}
	}()

	args := []string{"command", "arg1", "arg2"}
	data := "This is cached data."
	ttl := 1

	t.Run("SetAndValidateCacheFile", func(t *testing.T) {
		// Set cache
		err := Set(args, data, ttl)
		assert.NoError(t, err, "Failed to set cache")

		// Verify the cache file was created
		cacheKey := generateCacheKey(args)
		cacheFile := getCacheFileName(cacheKey)
		_, err = os.Stat(cacheFile)
		assert.NoError(t, err, "Cache file not created")

		// Read the file and verify contents
		bytes, err := ioutil.ReadFile(cacheFile)
		assert.NoError(t, err, "Failed to read cache file")
		assert.Contains(t, string(bytes), data, "Cache file does not contain expected data")
	})

	t.Run("ValidateCacheData", func(t *testing.T) {
		// Verify expiration (TTL)
		cachedData, found, err := Get(args)
		assert.True(t, found, "Cache entry not found")
		assert.NoError(t, err, "Failed to get cache")
		assert.Equal(t, data, cachedData, "Cached data does not match")
	})

	t.Run("ValidateExpirationTime", func(t *testing.T) {
		// Open the cache file manually and verify the expiration time
		cacheKey := generateCacheKey(args)
		cacheFile := getCacheFileName(cacheKey)
		file, _ := os.Open(cacheFile)
		defer file.Close()
		decoder := gob.NewDecoder(file)
		var cacheItem CacheItem
		_ = decoder.Decode(&cacheItem)
		assert.WithinDuration(t, cacheItem.Expiration, time.Now().Add(time.Duration(ttl)*time.Second), 1*time.Second, "Cache expiration does not match expected TTL")
	})
}

// cache_test.go (continuation)

func TestGet(t *testing.T) {
	args := []string{"command", "arg1", "arg2"}
	data := "This is cached data."
	ttl := 5 // 5 seconds

	// Setting cache to test Get
	err := Set(args, data, ttl)
	assert.NoError(t, err, "Failed to set cache for Get test")

	t.Run("GetExistingData", func(t *testing.T) {
		cachedData, found, err := Get(args)
		assert.True(t, found, "Cache entry not found")
		assert.NoError(t, err, "Failed to get cache")
		assert.Equal(t, data, cachedData, "Cached data does not match original data")
	})

	t.Run("GetNonExistentData", func(t *testing.T) {
		nonExistentArgs := []string{"command", "nonexistent"}
		_, found, err := Get(nonExistentArgs)
		assert.False(t, found, "Cache entry should not be found")
		assert.NoError(t, err, "There should be no error retrieving a non-existent cache entry")
	})

	t.Run("GetDataAfterExpiration", func(t *testing.T) {
		time.Sleep(time.Duration(ttl+1) * time.Second)
		_, found, err := Get(args)
		assert.False(t, found, "Cache entry should be expired and not found")
		assert.NoError(t, err, "There should be no error retrieving an expired cache entry")
	})

	// Cleanup after tests
	files, _ := filepath.Glob("/tmp/" + cachePrefix + "*.gob")
	for _, file := range files {
		os.Remove(file)
	}
}
