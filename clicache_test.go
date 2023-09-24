package clicache

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	// Cleanup utility to ensure no cache files are left after the test
	defer func() {
		files, _ := filepath.Glob("/tmp/" + cachePrefix + "*.gob")
		for _, file := range files {
			os.Remove(file)
		}
	}()

	args := []string{"command", "arg1", "arg2"}
	data := "This is cached data."
	ttl := 1

	// Set cache
	if err := Set(args, data, ttl); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Verify the cache file was created
	cacheKey := generateCacheKey(args)
	cacheFile := getCacheFileName(cacheKey)
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Fatalf("Cache file not created")
	}

	// Verify expiration (TTL)
	cachedData, found, err := Get(args)
	if !found {
		t.Fatal("Cache entry not found")
	}
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}
	if cachedData != data {
		t.Fatalf("Cached data does not match: got %v, want %v", cachedData, data)
	}

	// Open the cache file manually and verify the expiration time
	file, _ := os.Open(cacheFile)
	defer file.Close()
	decoder := gob.NewDecoder(file)
	var cacheItem CacheItem
	if err := decoder.Decode(&cacheItem); err != nil {
		t.Fatalf("Failed to decode cache item: %v", err)
	}
	if time.Now().Add(time.Duration(ttl)*time.Second).Sub(cacheItem.Expiration) > 1*time.Second {
		t.Fatalf("Cache expiration does not match expected TTL")
	}
}

func TestGet(t *testing.T) {
	args := []string{"command", "arg1", "arg2"}
	data := "This is cached data."
	ttl := 5 // 5 seconds

	// Setting cache to test Get
	if err := Set(args, data, ttl); err != nil {
		t.Fatalf("Failed to set cache for Get test: %v", err)
	}

	cachedData, found, err := Get(args)
	if !found {
		t.Fatal("Cache entry not found")
	}
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}
	if cachedData != data {
		t.Fatalf("Cached data does not match original data: got %v, want %v", cachedData, data)
	}

	nonExistentArgs := []string{"command", "nonexistent"}
	_, found, err = Get(nonExistentArgs)
	if found {
		t.Fatal("Cache entry should not be found")
	}
	if err != nil {
		t.Fatalf("There should be no error retrieving a non-existent cache entry: %v", err)
	}

	time.Sleep(time.Duration(ttl+1) * time.Second)
	_, found, err = Get(args)
	if found {
		t.Fatal("Cache entry should be expired and not found")
	}
	if err != nil {
		t.Fatalf("There should be no error retrieving an expired cache entry: %v", err)
	}

	// Cleanup after tests
	files, _ := filepath.Glob("/tmp/" + cachePrefix + "*.gob")
	for _, file := range files {
		os.Remove(file)
	}
}

func contains(haystack, needle string) bool {
	return filepath.HasPrefix(haystack, needle)
}
