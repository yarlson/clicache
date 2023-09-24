# CLICache Library

[![codecov](https://codecov.io/gh/yarlson/clicache/graph/badge.svg?token=2U3ILh24ya)](https://codecov.io/gh/yarlson/clicache)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

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
import "github.com/yarlson/cli-cache"

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

## Contributions

Contributions to clicache are welcome! Feel free to open issues or submit pull requests.

## License

clicache is licensed under the [MIT License](LICENSE).
