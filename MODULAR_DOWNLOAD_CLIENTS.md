# Modular Download Client Architecture

## Overview

The original `download_poller.go` was tightly coupled to qBittorrent and SABnzbd, making it difficult to add support for new download clients. This document outlines the new modular architecture that makes the system easily extensible.

## Problems with the Original System

1. **Tight Coupling**: The poller directly implemented qBittorrent and SABnzbd logic
2. **Hard to Extend**: Adding new clients required modifying core polling logic
3. **Code Duplication**: Similar logic was repeated for different clients
4. **Testing Difficulties**: Hard to test individual client implementations
5. **Maintenance Burden**: Changes to one client could affect others

## New Modular Architecture

### 1. Interface-Based Design

The new system uses a clean interface that all download clients must implement:

```go
type Interface interface {
    GetType() string
    GetName() string
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    GetDownloads(ctx context.Context) ([]Item, error)
    GetDownloadProgress(ctx context.Context, downloadID string) (*Progress, error)
    IsConnected() bool
    GetConnectionInfo() ConnectionInfo
}
```

### 2. Factory Pattern

A factory manages client registration and creation:

```go
type DownloadClientFactory struct {
    clients map[string]func(config downloadclient.Config) (downloadclient.Interface, error)
}
```

### 3. Client Manager

A manager handles multiple client connections:

```go
type DownloadClientManager struct {
    factory *DownloadClientFactory
    clients map[string]downloadclient.Interface
}
```

## File Structure

```
backend/
├── pkg/
│   └── downloadclient/
│       └── interface.go          # Core interface and types
├── internal/
│   └── integrations/
│       ├── download_clients.go   # Factory and manager
│       ├── modular_download_poller.go  # New modular poller
│       ├── download_poller.go    # Original poller (for comparison)
│       └── clients/
│           ├── init.go           # Client registration
│           ├── qbittorrent.go    # qBittorrent implementation
│           ├── sabnzbd.go        # SABnzbd implementation
│           ├── deluge.go         # Deluge placeholder
│           └── README.md         # Documentation for contributors
```

## Benefits of the New Architecture

### 1. **Easy Extension**
Adding a new client requires only:
- Implementing the interface
- Registering the client
- No changes to core polling logic

### 2. **Separation of Concerns**
- Each client handles its own connection and data conversion
- Core polling logic is independent of specific clients
- Clear boundaries between responsibilities

### 3. **Better Testing**
- Individual clients can be tested in isolation
- Mock clients can be easily created
- Integration tests can focus on the manager

### 4. **Improved Maintainability**
- Changes to one client don't affect others
- Common patterns can be shared
- Error handling is consistent across clients

### 5. **Type Safety**
- Strong typing prevents runtime errors
- Interface ensures all clients implement required methods
- Compile-time checking of client implementations

## Migration Strategy

The new system is designed to coexist with the original:

1. **Phase 1**: Implement new clients alongside existing ones
2. **Phase 2**: Test new clients thoroughly
3. **Phase 3**: Migrate to `ModularDownloadPoller` in production
4. **Phase 4**: Remove old `DownloadPoller` code

## Adding New Clients

### Example: Adding Transmission Support

1. **Create the client implementation**:
```go
// clients/transmission.go
type TransmissionClient struct {
    config     downloadclient.Config
    httpClient *http.Client
    connected  bool
    lastError  string
}

func NewTransmissionClient(config downloadclient.Config) (downloadclient.Interface, error) {
    return &TransmissionClient{config: config}, nil
}

// Implement all interface methods...
```

2. **Register the client**:
```go
// clients/init.go
func RegisterAll(factory interface {
    RegisterClient(clientType string, constructor func(config downloadclient.Config) (downloadclient.Interface, error))
}) {
    factory.RegisterClient("qbittorrent", NewQBitTorrentClient)
    factory.RegisterClient("sabnzbd", NewSABnzbdClient)
    factory.RegisterClient("deluge", NewDelugeClient)
    factory.RegisterClient("transmission", NewTransmissionClient) // New line
}
```

3. **Update database schema** (if needed):
```sql
ALTER TABLE download_clients 
ADD CONSTRAINT download_clients_type_check 
CHECK (type IN ('qbittorrent', 'sabnzbd', 'deluge', 'transmission'));
```

That's it! The new client is now supported without any changes to the core polling logic.

## Supported Clients

### Currently Implemented
- **qBittorrent**: Full implementation with session management
- **SABnzbd**: Full implementation with API key authentication
- **Deluge**: Placeholder implementation (ready for completion)

### Easy to Add
- **Transmission**: RPC-based, well-documented API
- **rTorrent**: XML-RPC interface
- **uTorrent**: Web API
- **NZBGet**: Similar to SABnzbd
- **Any custom client**: Just implement the interface

## Performance Considerations

### Connection Management
- Clients maintain persistent connections where possible
- Automatic reconnection on failures
- Connection pooling for high-traffic scenarios

### Caching
- Client-specific caching can be implemented
- Progress data cached to reduce API calls
- Configurable cache invalidation

### Error Handling
- Circuit breakers prevent cascading failures
- Graceful degradation when clients are unavailable
- Detailed error reporting for debugging

## Future Enhancements

### 1. **Plugin System**
Allow external plugins to register new clients without modifying core code.

### 2. **Configuration Validation**
Client-specific configuration validation and documentation.

### 3. **Metrics and Monitoring**
Per-client metrics for performance monitoring.

### 4. **Load Balancing**
Support for multiple instances of the same client type.

### 5. **Advanced Matching**
Better algorithms for matching downloads with Radarr/Sonarr data.

## Conclusion

The new modular architecture transforms the download client system from a tightly coupled, hard-to-extend implementation into a flexible, maintainable, and easily extensible system. Contributors can now add support for new download clients with minimal effort and without risk of breaking existing functionality.

The architecture follows Go best practices and provides a solid foundation for future enhancements while maintaining backward compatibility with the existing system. 