# Download Client System

This directory contains the modular download client implementations for the Serra application. The system is designed to be easily extensible, allowing contributors to add support for new download clients without modifying the core polling logic.

## Architecture

The download client system uses an interface-based architecture with the following components:

### Core Interface (`pkg/downloadclient/interface.go`)

The `downloadclient.Interface` defines the contract that all download clients must implement:

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

### Factory Pattern (`internal/integrations/download_clients.go`)

The `DownloadClientFactory` manages client registration and creation:

```go
type DownloadClientFactory struct {
    clients map[string]func(config downloadclient.Config) (downloadclient.Interface, error)
}
```

### Client Manager (`internal/integrations/download_clients.go`)

The `DownloadClientManager` handles multiple client connections and provides a unified interface for retrieving downloads from all configured clients.

## Adding a New Download Client

To add support for a new download client (e.g., Transmission, rTorrent, etc.), follow these steps:

### 1. Create the Client Implementation

Create a new file in the `clients/` directory (e.g., `transmission.go`):

```go
package clients

import (
    "context"
    "fmt"
    "github.com/mahcks/serra/pkg/downloadclient"
)

type TransmissionClient struct {
    config     downloadclient.Config
    connected  bool
    lastError  string
    // Add any client-specific fields
}

func NewTransmissionClient(config downloadclient.Config) (downloadclient.Interface, error) {
    client := &TransmissionClient{
        config: config,
    }
    return client, nil
}

// Implement all required interface methods
func (c *TransmissionClient) GetType() string {
    return "transmission"
}

func (c *TransmissionClient) GetName() string {
    return c.config.Name
}

func (c *TransmissionClient) Connect(ctx context.Context) error {
    // Implement connection logic
    // - Establish connection to Transmission
    // - Authenticate if required
    // - Set connected = true on success
    return nil
}

func (c *TransmissionClient) Disconnect(ctx context.Context) error {
    // Implement disconnection logic
    c.connected = false
    return nil
}

func (c *TransmissionClient) GetDownloads(ctx context.Context) ([]downloadclient.Item, error) {
    // Implement download retrieval logic
    // - Call Transmission API to get torrent list
    // - Convert to []downloadclient.Item format
    // - Return only active downloads
    return nil, nil
}

func (c *TransmissionClient) GetDownloadProgress(ctx context.Context, downloadID string) (*downloadclient.Progress, error) {
    // Implement progress retrieval logic
    // - Call Transmission API to get specific torrent info
    // - Convert to downloadclient.Progress format
    return nil, nil
}

func (c *TransmissionClient) IsConnected() bool {
    return c.connected
}

func (c *TransmissionClient) GetConnectionInfo() downloadclient.ConnectionInfo {
    return downloadclient.ConnectionInfo{
        Host:      c.config.Host,
        Port:      c.config.Port,
        UseSSL:    c.config.UseSSL,
        Connected: c.connected,
        LastError: c.lastError,
    }
}
```

### 2. Register the Client

Add the client to the registration in `clients/init.go`:

```go
func RegisterAll(factory interface {
    RegisterClient(clientType string, constructor func(config downloadclient.Config) (downloadclient.Interface, error))
}) {
    factory.RegisterClient("qbittorrent", NewQBitTorrentClient)
    factory.RegisterClient("sabnzbd", NewSABnzbdClient)
    factory.RegisterClient("deluge", NewDelugeClient)
    factory.RegisterClient("transmission", NewTransmissionClient) // Add this line
}
```

### 3. Update Database Schema (if needed)

If the new client requires additional configuration fields, update the database schema in `internal/db/schema.sql`:

```sql
-- Example: Add new client type to the CHECK constraint
ALTER TABLE download_clients 
DROP CONSTRAINT download_clients_type_check;

ALTER TABLE download_clients 
ADD CONSTRAINT download_clients_type_check 
CHECK (type IN ('qbittorrent', 'sabnzbd', 'deluge', 'transmission'));
```

### 4. Test the Implementation

Create tests for your new client implementation:

```go
// clients/transmission_test.go
package clients

import (
    "context"
    "testing"
    "github.com/mahcks/serra/pkg/downloadclient"
)

func TestTransmissionClient(t *testing.T) {
    config := downloadclient.Config{
        ID:   "test-transmission",
        Type: "transmission",
        Name: "Test Transmission",
        Host: "localhost",
        Port: 9091,
    }

    client, err := NewTransmissionClient(config)
    if err != nil {
        t.Fatalf("Failed to create client: %v", err)
    }

    // Test connection
    err = client.Connect(context.Background())
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }

    // Test getting downloads
    downloads, err := client.GetDownloads(context.Background())
    if err != nil {
        t.Fatalf("Failed to get downloads: %v", err)
    }

    // Verify downloads format
    for _, download := range downloads {
        if download.ID == "" {
            t.Error("Download ID should not be empty")
        }
        if download.Progress < 0 || download.Progress > 100 {
            t.Error("Progress should be between 0 and 100")
        }
    }

    // Test disconnection
    err = client.Disconnect(context.Background())
    if err != nil {
        t.Fatalf("Failed to disconnect: %v", err)
    }
}
```

## Best Practices

### 1. Error Handling

- Always set `lastError` when operations fail
- Return meaningful error messages
- Handle connection timeouts gracefully
- Implement retry logic for transient failures

### 2. Resource Management

- Properly close connections in `Disconnect()`
- Use context cancellation for long-running operations
- Implement connection pooling if needed

### 3. Data Conversion

- Convert client-specific data formats to the standard `downloadclient.Item` format
- Handle missing or null values gracefully
- Normalize progress values to 0-100 range
- Format time values consistently

### 4. Logging

- Use structured logging with appropriate log levels
- Include relevant context (client type, host, port)
- Log connection events and errors

### 5. Configuration

- Validate required configuration fields
- Support both authentication methods (username/password and API keys)
- Handle SSL/TLS configuration properly

## Example: Transmission Client Implementation

Here's a more complete example of how to implement a Transmission client:

```go
package clients

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
    "github.com/mahcks/serra/pkg/downloadclient"
)

type TransmissionClient struct {
    config     downloadclient.Config
    httpClient *http.Client
    sessionID  string
    connected  bool
    lastError  string
}

type transmissionRequest struct {
    Method    string      `json:"method"`
    Arguments interface{} `json:"arguments,omitempty"`
    Tag       int         `json:"tag"`
}

type transmissionResponse struct {
    Arguments transmissionArguments `json:"arguments"`
    Result    string               `json:"result"`
}

type transmissionArguments struct {
    Torrents []transmissionTorrent `json:"torrents"`
}

type transmissionTorrent struct {
    ID           int     `json:"id"`
    Name         string  `json:"name"`
    HashString   string  `json:"hashString"`
    PercentDone  float64 `json:"percentDone"`
    Status       int     `json:"status"`
    SizeWhenDone int64   `json:"sizeWhenDone"`
    LeftUntilDone int64  `json:"leftUntilDone"`
    Eta          int     `json:"eta"`
    RateDownload int64   `json:"rateDownload"`
    RateUpload   int64   `json:"rateUpload"`
    DateAdded    int64   `json:"dateAdded"`
}

func NewTransmissionClient(config downloadclient.Config) (downloadclient.Interface, error) {
    return &TransmissionClient{
        config: config,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }, nil
}

func (c *TransmissionClient) Connect(ctx context.Context) error {
    // Test connection by getting session ID
    req := transmissionRequest{
        Method: "session-get",
        Tag:    1,
    }
    
    _, err := c.makeRequest(ctx, req)
    if err != nil {
        c.lastError = err.Error()
        return err
    }
    
    c.connected = true
    c.lastError = ""
    return nil
}

func (c *TransmissionClient) GetDownloads(ctx context.Context) ([]downloadclient.Item, error) {
    if !c.connected {
        return nil, fmt.Errorf("not connected to Transmission")
    }
    
    req := transmissionRequest{
        Method: "torrent-get",
        Arguments: map[string]interface{}{
            "fields": []string{
                "id", "name", "hashString", "percentDone", "status",
                "sizeWhenDone", "leftUntilDone", "eta", "rateDownload",
                "rateUpload", "dateAdded",
            },
        },
        Tag: 2,
    }
    
    resp, err := c.makeRequest(ctx, req)
    if err != nil {
        return nil, err
    }
    
    var downloads []downloadclient.Item
    for _, torrent := range resp.Arguments.Torrents {
        // Only include active downloads (status 4 = downloading, 6 = seeding)
        if torrent.Status == 4 || torrent.Status == 6 {
            download := downloadclient.Item{
                ID:          strconv.Itoa(torrent.ID),
                Name:        torrent.Name,
                Hash:        torrent.HashString,
                Progress:    torrent.PercentDone * 100,
                Status:      c.getStatusString(torrent.Status),
                Size:        torrent.SizeWhenDone,
                SizeLeft:    torrent.LeftUntilDone,
                TimeLeft:    c.formatTimeLeft(torrent.Eta),
                DownloadSpeed: c.formatSpeed(torrent.RateDownload),
                UploadSpeed:   c.formatSpeed(torrent.RateUpload),
                ETA:          int64(torrent.Eta),
                AddedOn:      time.Unix(torrent.DateAdded, 0),
            }
            downloads = append(downloads, download)
        }
    }
    
    return downloads, nil
}

func (c *TransmissionClient) makeRequest(ctx context.Context, req transmissionRequest) (*transmissionResponse, error) {
    // Implementation of HTTP request to Transmission RPC
    // This would include authentication and session management
    return nil, nil
}

func (c *TransmissionClient) getStatusString(status int) string {
    switch status {
    case 0:
        return "stopped"
    case 1:
        return "check_wait"
    case 2:
        return "check"
    case 3:
        return "download_wait"
    case 4:
        return "downloading"
    case 5:
        return "seed_wait"
    case 6:
        return "seeding"
    default:
        return "unknown"
    }
}

// ... implement other required methods
```

## Migration from Old System

The new modular system is designed to be backward compatible. The existing `DownloadPoller` can be gradually replaced with the `ModularDownloadPoller`:

1. **Phase 1**: Implement new clients alongside existing ones
2. **Phase 2**: Test new clients thoroughly
3. **Phase 3**: Migrate to `ModularDownloadPoller` in production
4. **Phase 4**: Remove old `DownloadPoller` code

## Contributing

When contributing a new download client:

1. Follow the existing code style and patterns
2. Include comprehensive tests
3. Document any client-specific configuration requirements
4. Update this README with any new patterns or best practices
5. Consider edge cases and error conditions
6. Ensure the implementation is efficient and doesn't impact performance

## Support

For questions about implementing new download clients or issues with the modular system, please:

1. Check the existing client implementations for examples
2. Review the interface documentation in `pkg/downloadclient/interface.go`
3. Look at the test files for usage patterns
4. Open an issue with specific questions or problems 