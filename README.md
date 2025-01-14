# GeeCache

GeeCache is a distributed caching system implemented in Go, inspired by GroupCache. It provides a concurrent, distributed key-value cache with automatic node discovery and load balancing capabilities.

## Features

- LRU (Least Recently Used) cache implementation
- Concurrent access support using mutexes
- HTTP-based peer-to-peer communication
- Consistent hashing for distributed node management
- Group caching with automatic data loading
- Thread-safe operations

## Project Structure

- `/cache`: Core cache implementation with thread-safe operations
- `/lru`: LRU cache implementation
- `/http`: HTTP server and client for peer communication, includes consistent hashing
- `/group`: Main cache control and coordination logic

## Usage

```go
import "geecache"

// Create a new cache group
group := geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
    func(key string) ([]byte, error) {
        // Your data loading logic here
        return []byte(value), nil
    }))

// Get value from cache
value, err := group.Get("key")
```

## Components

### Cache Layer
The cache package provides the core caching functionality with thread-safe operations and memory management.

### LRU Implementation
The LRU package implements the Least Recently Used cache eviction policy, ensuring efficient memory usage.

### HTTP Layer
Handles peer-to-peer communication between cache nodes, implementing both server and client functionalities.

### Consistent Hashing
Implements consistent hashing for distributed node management, ensuring even distribution of cache entries across nodes.

### Group
Manages cache groups and coordinates between different components, handling cache misses and data loading.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
