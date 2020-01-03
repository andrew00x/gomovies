package service

import (
	"testing"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/gomovies/pkg/torrent"
	"github.com/stretchr/testify/assert"
)

var tr *torrentMock

func TestAddFile(t *testing.T) {
	setup()
	srv := &TorrentService{tr: tr}

	err := srv.AddFile([]byte("torrent"))

	assert.Nil(t, err)
	assert.Equal(t, []interface{}{[]byte("torrent")}, tr.added)
}

func TestAddUrl(t *testing.T) {
	setup()
	srv := &TorrentService{tr: tr}

	err := srv.AddUrl("torrent")

	assert.Nil(t, err)
	assert.Equal(t, []interface{}{"torrent"}, tr.added)
}

func TestGetTorrents(t *testing.T) {
	setup()
	tr.downloads = []api.TorrentDownload{{Name: "foo", Completed: true}, {Name: "bar", Completed: true}}
	srv := &TorrentService{tr: tr}

	res, err := srv.Torrents()
	assert.Nil(t, err)
	assert.Equal(t, []api.TorrentDownload{{Name: "foo", Completed: true}, {Name: "bar", Completed: true}}, res)
}

func TestStopTorrent(t *testing.T) {
	setup()
	srv := &TorrentService{tr: tr}

	err := srv.Stop(api.TorrentDownload{Name: "foo"})
	assert.Nil(t, err)
	assert.Equal(t, []api.TorrentDownload{{Name: "foo"}}, tr.stopped)
}

func TestStartTorrent(t *testing.T) {
	setup()
	srv := &TorrentService{tr: tr}

	err := srv.Start(api.TorrentDownload{Name: "foo"})
	assert.Nil(t, err)
	assert.Equal(t, []api.TorrentDownload{{Name: "foo"}}, tr.started)
}

func TestDeleteTorrent(t *testing.T) {
	setup()
	srv := &TorrentService{tr: tr}

	err := srv.Delete(api.TorrentDownload{Name: "foo"})
	assert.Nil(t, err)
	assert.Equal(t, []api.TorrentDownload{{Name: "foo"}}, tr.deleted)
}

func setup() {
	tr = &torrentMock{}
	torrent.Factory = func(*config.Config) torrent.Torrent {
		return tr
	}
}

type torrentMock struct {
	err           error
	downloads     []api.TorrentDownload
	added         []interface{}
	deleted       []api.TorrentDownload
	started       []api.TorrentDownload
	stopped       []api.TorrentDownload
}

func (t *torrentMock) AddFile(b []byte) error {
	t.added = append(t.added, b)
	return t.err
}

func (t *torrentMock) AddUrl(u string) error {
	t.added = append(t.added, u)
	return t.err
}

func (t *torrentMock) Torrents() ([]api.TorrentDownload, error) {
	return t.downloads, t.err
}

func (t *torrentMock) Stop(d api.TorrentDownload) error {
	t.stopped = append(t.stopped, d)
	return t.err
}

func (t *torrentMock) Start(d api.TorrentDownload) error {
	t.started = append(t.started, d)
	return t.err
}

func (t *torrentMock) Delete(d api.TorrentDownload) error {
	t.deleted = append(t.deleted, d)
	return t.err
}
