LazyFetcher is a Go package that implements a lazy-loading, expiring cache for fetching and storing generic data.

## Features

- Generic implementation for flexible data types
- Lazy loading of data
- Automatic expiration and refresh of cached entries
- Thread-safe operations using `sync.Map`
- Support for prioritized data retrieval

This package is ideal for scenarios where you need to cache data with automatic expiration and refresh capabilities, especially in high-concurrency environments.

## Installation

```
go get github.com/afjoseph/lazyfetcher
```

## Example

```
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/afjoseph/lazyfetcher"
)

func main() {
    // Create a fetcher function
    fetchFunc := func(ctx context.Context, key string) ([]string, int, error) {
        // Simulate an expensive operation
        time.Sleep(time.Second)
        return []string{key + "-value1", key + "-value2"}, 0, nil
    }

    // Create a new LazyFetcher instance
    fetcher := lazyfetcher.New[string](time.Minute, fetchFunc)

    // Fetch data
    ctx := context.Background()
    values, highestPriorityIdx, _ := fetcher.Fetch(ctx, "example-key")

    fmt.Printf("Values: %v, Highest Priority Index: %d\n", values, highestPriorityIdx)

    // Fetch again (this time it should be cached)
    values, highestPriorityIdx, _ = fetcher.Fetch(ctx, "example-key")
    // .FetchPriority() is a shortcut for values[highestPriorityIdx]

    fmt.Printf("Cached Values: %v, Highest Priority Index: %d\n", values, highestPriorityIdx)
}
```
