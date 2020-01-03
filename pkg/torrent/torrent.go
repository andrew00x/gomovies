package torrent

import (
	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
)

type factory func(*config.Config) Torrent

var Factory factory

func CreateTorrent(cfg *config.Config) Torrent {
	return Factory(cfg)
}

type Torrent interface {
	AddFile([]byte) error
	AddUrl(string) error
	Torrents() ([]api.TorrentDownload, error)
	Start(api.TorrentDownload) error
	Stop(api.TorrentDownload) error
	Delete(api.TorrentDownload) error
}
