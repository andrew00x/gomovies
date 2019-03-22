package service

import (
	"sync"
	"time"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/gomovies/pkg/torrent"
)

type TorrentService struct {
	mu                      sync.Mutex
	tr                      torrent.Torrent
	idleTime                time.Duration
	idle                    bool
	idleTorrentClientTicker *time.Ticker
	conf                    *config.Config
}

func CreateTorrentService(conf *config.Config) *TorrentService {
	return &TorrentService{tr: torrent.CreateTorrent(conf)}
}

func (srv *TorrentService) AddFile(file []byte) error {
	return srv.tr.AddFile(file)
}

func (srv *TorrentService) AddUrl(url string) error {
	return srv.tr.AddUrl(url)
}

func (srv *TorrentService) Torrents() ([]api.TorrentDownload, error) {
	return srv.tr.Torrents()
}

func (srv *TorrentService) Files(d api.TorrentDownload) ([]api.TorrentDownloadFile, error) {
	return srv.tr.Files(d)
}

func (srv *TorrentService) Stop(d api.TorrentDownload) error {
	return srv.tr.Stop(d)
}

func (srv *TorrentService) Start(d api.TorrentDownload) error {
	return srv.tr.Start(d)
}

func (srv *TorrentService) Delete(d api.TorrentDownload) error {
	return srv.tr.Delete(d)
}
