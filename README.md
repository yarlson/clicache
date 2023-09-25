# CLICache Library

[![codecov](https://codecov.io/gh/yarlson/clicache/graph/badge.svg?token=2U3ILh24ya)](https://codecov.io/gh/yarlson/clicache)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub tag checks state](https://img.shields.io/github/checks-status/yarlson/clicache/main)
[![Go Reference](https://pkg.go.dev/badge/github.com/yarlson/clicache.svg)](https://pkg.go.dev/github.com/yarlson/clicache)

`clicache` is a lightweight Go library designed to provide simple file-based caching for Command Line Interface (CLI)
applications. By leveraging the local filesystem, it allows quick storage and retrieval of data between CLI invocations,
making it useful for operations that don't need to compute or fetch data repeatedly within a short time span.

## Features

- **File-Based Caching**: Store cache data directly on the filesystem in the `/tmp` directory.
- **TTL Support**: Set an expiration time for cached data.
- **Automatic Cleanup**: Garbage collection to automatically remove expired cache entries.
- **Concurrency Safe**: Uses locks to ensure safe concurrent access.

## Installation

To install `clicache`, use `go get`:

```bash
go get github.com/yarlson/clicache@v0.2.0
```

## Usage

Here are the basic operations that clicache supports:

### Setting Cache Data

Store data in the cache with a specific set of command arguments and a TTL (Time-to-Live) in seconds.

```go
package main

import "github.com/yarlson/clicache"

func main() {
    args := []string{"my-command", "arg1", "arg2"}
    data := "This is some data to cache."
    ttl := 60 // Cache for 60 seconds
    
    err := clicache.Set(args, data, ttl)
    if err != nil {
    // Handle error
    }
}
```

### Getting Cache Data

Retrieve data from the cache using a specific set of command arguments.

```go
package main

import (
    "fmt"
    "github.com/yarlson/clicache"
)

func main() {
	args := []string{"my-command", "arg1", "arg2"}

	data, found, err := clicache.Get(args)
	if err != nil {
		// Handle error
	}
	if found {
		// Use the data
		fmt.Println(data)
	}
}

```

### Using the Cache Helper Function

The `Cache` function provides a convenient way to get cached data based on provided CLI arguments. If the data is not found in the cache, the function defined in the handler is executed and its result is then cached with the specified TTL.

```go
package main

import (
	"fmt"
	"github.com/yarlson/clicache"
)

func main() {
	out, err := clicache.Cache(func() (string, error) {
		// This function is only executed if the data is not in the cache.
		return "This is data.", nil
	})

	if err != nil {
		// Handle error
	}
	fmt.Println(out)  // This will print "This is data."
}
```

### Setting Default TTL for Cache Entries

You can set a default Time-to-Live (TTL) in seconds for cache entries using the `SetTTL` function. This TTL value will be applied to all subsequent cache entries unless specifically overridden during the cache set operation.

```go
package main

import "github.com/yarlson/clicache"

func main() {
    // Set the default TTL to 1 minute
    clicache.SetTTL(60)

    // Other operations using clicache can follow
    // ...
}

```

### Cache Garbage Collection

The `GC` function allows for manual cleanup of the cache directory by scanning for and removing expired cache entries. This ensures that the cache stays efficient and doesn't accumulate outdated data.

```go
package main

import "github.com/yarlson/clicache"

func main() {
    // Perform garbage collection to remove outdated cache entries
    clicache.GC()

    // Other operations using clicache can follow
    // ...
}

```

## Contributions

Contributions to clicache are welcome! Feel free to open issues or submit pull requests.

## License

clicache is licensed under the [MIT License](LICENSE).
