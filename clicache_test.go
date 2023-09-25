package clicache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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

func TestSet(t *testing.T) {
	type args struct {
		args []string
		data interface{}
		ttl  int
	}
	tests := []struct {
		name    string
		args    args
		fs      FileSystem
		wantErr bool
	}{
		{
			name: "Happy path",
			args: args{
				args: []string{"command", "arg1", "arg2"},
				data: "This is cached data.",
				ttl:  1,
			},
			fs:      fs,
			wantErr: false,
		},
		{
			name: "Cannot create cache file",
			args: args{
				args: []string{"../../../command", "arg1", "arg2"},
				data: "This is cached data.",
				ttl:  1,
			},
			fs: &FileSystemMock{
				CreateFunc: func(name string) (*os.File, error) {
					return nil, errors.New("error")
				},
			},
			wantErr: true,
		},
		{
			name: " IsNotExist error",
			args: args{
				args: []string{"command", "arg1", "arg2"},
				data: "This is cached data.",
				ttl:  1,
			},
			fs: &FileSystemMock{
				CreateFunc: func(name string) (*os.File, error) {
					f, _ := os.Create(getCacheFileName(name))
					return f, nil
				},
				OpenFunc: func(name string) (*os.File, error) {
					return nil, errors.New("error")
				},
				IsNotExistFunc: func(err error) bool {
					return false
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs = tt.fs
			if err := Set(tt.args.args, tt.args.data, tt.args.ttl); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetTTL(t *testing.T) {
	type args struct {
		ttl int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Happy path",
			args: args{
				ttl: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTTL(tt.args.ttl)
			if cacheTTL != tt.args.ttl {
				t.Errorf("SetTTL() = %v, want %v", cacheTTL, tt.args.ttl)
			}
		})
	}
}
