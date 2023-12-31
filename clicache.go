// Package clicache provides file-based caching tailored for CLI applications.
// It allows CLI applications to cache data based on command arguments, and
// supports TTL-based cache expiration.
package clicache

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileSystem is an interface for file system operations.
//
//go:generate moq -skip-ensure -out fs_mock_test.go -fmt goimports . FileSystem
type FileSystem interface {
	Create(name string) (*os.File, error)
	Open(name string) (*os.File, error)
	Remove(name string) error
	IsNotExist(err error) bool
}

// OSFileSystem is an implementation of FileSystem that uses the OS file system.
type OSFileSystem struct{}

func (o OSFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (o OSFileSystem) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (o OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (o OSFileSystem) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// fs is the file system used by clicache.
var fs FileSystem = OSFileSystem{}

// CacheItem represents a cached item with its expiration time and data.
type CacheItem struct {
	Expiration time.Time
	Data       interface{}
}

var (
	cacheMutex  sync.Mutex
	cachePrefix = "cli_cache_"
	cacheTTL    = 300
	cacheFolder = "/tmp/"
)

// SetTTL sets the default TTL for cache entries.
//
// ttl: Time to live in seconds for the cache entry.
//
// Example:
//
//	clicache.SetTTL(60)  // 1 minute
func SetTTL(ttl int) {
	cacheTTL = ttl
}

// generateCacheKey produces a unique cache key based on the provided CLI arguments.
// This ensures that different command invocations have distinct cache entries.
func generateCacheKey(args []string) string {
	joinedArgs := fmt.Sprintf("%v", args)
	hash := sha256.Sum256([]byte(joinedArgs))
	return hex.EncodeToString(hash[:])
}

// getCacheFileName constructs the cache file name for the given cache key.
func getCacheFileName(cacheKey string) string {
	return filepath.Join(cacheFolder, cachePrefix+fmt.Sprintf("%s.gob", cacheKey))
}

// Cache is a helper function that retrieves the cached data associated with the provided CLI arguments.
// If the cache entry is not found, the provided handler function is executed and its output is cached.
// The data will expire after the specified TTL (in seconds).
//
// handler: Function that returns the data to be cached.
//
// Returns the cached data and an error if the operation fails.
//
// Example:
//
//	out, err := clicache.Cache(func() (string, error) {
//	  return "This is data.", nil
//	})
func Cache(handler func() (string, error)) (string, error) {
	cached, isCached, err := Get(flag.Args())
	if err != nil {
		return "", err
	}
	if isCached {
		return cached.(string), nil
	}

	out, err := handler()
	if err != nil {
		return "", err
	}

	err = Set(flag.Args(), out, cacheTTL)
	if err != nil {
		return "", err
	}

	return out, nil
}

// Set stores the given data in the cache, associated with the provided CLI arguments.
// The data will expire after the specified TTL (in seconds).
//
// args: Command line arguments which determine the cache key.
// data: Data to be cached.
// ttl: Time to live in seconds for the cache entry.
//
// Returns an error if the operation fails.
//
// Example:
//
//	args := []string{"command", "arg1", "arg2"}
//	data := "This is cached data."
//	ttl := 60  // 1 minute
//	err := clicache.Set(args, data, ttl)
//	if err != nil {
//	  log.Fatalf("Failed to set cache: %v", err)
//	}
func Set(args []string, data interface{}, ttl int) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	cacheKey := generateCacheKey(args)
	cacheFile := getCacheFileName(cacheKey)
	cacheItem := CacheItem{
		Expiration: time.Now().Add(time.Duration(ttl) * time.Second),
		Data:       data,
	}

	file, err := fs.Create(cacheFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(&cacheItem)
	if err != nil {
		return err
	}

	gc() // Clean up expired cache entries.

	return nil
}

// Get retrieves the cached data associated with the provided CLI arguments.
//
// args: Command line arguments which determine the cache key.
//
// Returns the cached data, a boolean indicating if the cache entry was found, and an error if the operation fails.
//
// Example:
//
//	args := []string{"command", "arg1", "arg2"}
//	data, found, err := clicache.Get(args)
//	if err != nil {
//	  log.Fatalf("Failed to get cache: %v", err)
//	}
//	if found {
//	  fmt.Println("Cached data:", data)
//	} else {
//	  fmt.Println("Cache not found.")
//	}
func Get(args []string) (interface{}, bool, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	cacheKey := generateCacheKey(args)
	cacheFile := getCacheFileName(cacheKey)

	file, err := fs.Open(cacheFile)
	if err != nil {
		if fs.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var cacheItem CacheItem
	err = decoder.Decode(&cacheItem)

	gc() // Clean up expired cache entries.

	if err != nil || time.Now().After(cacheItem.Expiration) {
		_ = fs.Remove(cacheFile)
		return nil, false, nil
	}

	return cacheItem.Data, true, nil
}

// gc scans the cache directory and removes outdated cache entries.
// This ensures the cache stays lean and doesn't hoard expired data.
func gc() {
	files, err := filepath.Glob(cacheFolder + cachePrefix + "*.gob")
	if err != nil {
		return
	}

	for _, file := range files {
		f, err := fs.Open(file)
		if err != nil {
			continue
		}

		decoder := gob.NewDecoder(f)
		var cacheItem CacheItem
		err = decoder.Decode(&cacheItem)
		_ = f.Close()

		if err != nil || time.Now().After(cacheItem.Expiration) {
			_ = fs.Remove(file)
		}
	}
}

// Cleanup removes all cache entries.
//
// Example:
//
//	clicache.Cleanup()
func Cleanup() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	files, err := filepath.Glob(cacheFolder + cachePrefix + "*.gob")
	if err != nil {
		return
	}

	for _, file := range files {
		_, err := fs.Open(file)
		if err != nil {
			continue
		}

		_ = fs.Remove(file)
	}
}
