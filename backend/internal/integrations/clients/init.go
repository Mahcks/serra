package clients

import (
	"github.com/mahcks/serra/pkg/downloadclient"
)

// RegisterAll registers all available download clients with the factory
func RegisterAll(factory interface {
	RegisterClient(clientType string, constructor func(config downloadclient.Config) (downloadclient.Interface, error))
}) {
	factory.RegisterClient("qbittorrent", NewQBitTorrentClient)
	factory.RegisterClient("sabnzbd", NewSABnzbdClient)
}
