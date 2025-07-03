package integrations

import (
	"context"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/integrations/clients"
	"github.com/mahcks/serra/pkg/downloadclient"
	"github.com/mahcks/serra/utils"
)

// DownloadClientFactory creates download client instances
type DownloadClientFactory struct {
	clients map[string]func(config downloadclient.Config) (downloadclient.Interface, error)
}

// NewDownloadClientFactory creates a new factory instance
func NewDownloadClientFactory() *DownloadClientFactory {
	factory := &DownloadClientFactory{
		clients: make(map[string]func(config downloadclient.Config) (downloadclient.Interface, error)),
	}

	// Register all available clients
	clients.RegisterAll(factory)

	return factory
}

// RegisterClient registers a new download client type
func (f *DownloadClientFactory) RegisterClient(clientType string, constructor func(config downloadclient.Config) (downloadclient.Interface, error)) {
	f.clients[clientType] = constructor
}

// CreateClient creates a download client instance from database configuration
func (f *DownloadClientFactory) CreateClient(dbClient repository.DownloadClient) (downloadclient.Interface, error) {
	constructor, exists := f.clients[dbClient.Type]
	if !exists {
		return nil, &downloadclient.UnsupportedClientError{ClientType: dbClient.Type}
	}

	config := downloadclient.Config{
		ID:     dbClient.ID,
		Type:   dbClient.Type,
		Name:   dbClient.Name,
		Host:   dbClient.Host,
		Port:   int(dbClient.Port),
		UseSSL: utils.NullableBool{NullBool: dbClient.UseSsl}.Or(false),
	}

	config.Username = utils.NullableString{NullString: dbClient.Username}.ToPointer()
	config.Password = utils.NullableString{NullString: dbClient.Password}.ToPointer()
	config.APIKey = utils.NullableString{NullString: dbClient.ApiKey}.ToPointer()

	return constructor(config)
}

// GetSupportedClients returns a list of supported client types
func (f *DownloadClientFactory) GetSupportedClients() []string {
	clients := make([]string, 0, len(f.clients))
	for clientType := range f.clients {
		clients = append(clients, clientType)
	}
	return clients
}

// DownloadClientManager manages multiple download client connections
type DownloadClientManager struct {
	factory *DownloadClientFactory
	clients map[string]downloadclient.Interface
}

// NewDownloadClientManager creates a new client manager
func NewDownloadClientManager() *DownloadClientManager {
	return &DownloadClientManager{
		factory: NewDownloadClientFactory(),
		clients: make(map[string]downloadclient.Interface),
	}
}

// InitializeClients initializes all configured download clients
func (m *DownloadClientManager) InitializeClients(dbClients []repository.DownloadClient) error {
	for _, dbClient := range dbClients {
		client, err := m.factory.CreateClient(dbClient)
		if err != nil {
			return err
		}

		// Connect to the client
		if err := client.Connect(context.Background()); err != nil {
			return err
		}

		m.clients[dbClient.ID] = client
	}

	return nil
}

// GetClient returns a specific client by ID
func (m *DownloadClientManager) GetClient(clientID string) (downloadclient.Interface, bool) {
	client, exists := m.clients[clientID]
	return client, exists
}

// GetAllDownloads retrieves downloads from all connected clients
func (m *DownloadClientManager) GetAllDownloads(ctx context.Context) ([]downloadclient.Item, error) {
	var allDownloads []downloadclient.Item

	for clientID, client := range m.clients {
		if !client.IsConnected() {
			continue
		}

		downloads, err := client.GetDownloads(ctx)
		if err != nil {
			// Log error but continue with other clients
			continue
		}

		// Add client ID to each download for identification
		for i := range downloads {
			downloads[i].ID = clientID + "_" + downloads[i].ID
		}

		allDownloads = append(allDownloads, downloads...)
	}

	return allDownloads, nil
}

// GetDownloadProgress retrieves progress for a specific download
func (m *DownloadClientManager) GetDownloadProgress(ctx context.Context, downloadID string) (*downloadclient.Progress, error) {
	// Parse client ID from download ID (format: clientID_downloadID)
	// This is a simple implementation - you might want to make this more robust
	for clientID, client := range m.clients {
		if !client.IsConnected() {
			continue
		}

		// Check if this download belongs to this client
		if len(downloadID) > len(clientID)+1 && downloadID[:len(clientID)+1] == clientID+"_" {
			actualDownloadID := downloadID[len(clientID)+1:]
			return client.GetDownloadProgress(ctx, actualDownloadID)
		}
	}

	return nil, &downloadclient.DownloadNotFoundError{DownloadID: downloadID}
}

// CloseAll closes all client connections
func (m *DownloadClientManager) CloseAll(ctx context.Context) error {
	for _, client := range m.clients {
		if err := client.Disconnect(ctx); err != nil {
			return err
		}
	}
	return nil
}
