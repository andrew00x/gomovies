package torrent

import (
	"testing"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/xmlrpc/pkg/xmlrpc"
	"github.com/stretchr/testify/assert"
)

func TestCreateRtorrent(t *testing.T) {
	cfg := &config.Config{TorrentRemoteCtrlAddr: "/tmp/rtorrent.sock"}
	rt := createRtorrent(cfg).(*rtorrent)
	assert.Equal(t, "/tmp/rtorrent.sock", rt.rpc.(*xmlrpc.SCGIXmlRpc).Addr)
}

func TestAddTorrentUrl(t *testing.T) {
	rt := &rtorrent{}

	rpc := xmlRpcMock{}
	rt.rpc = &rpc
	err := rt.AddUrl("/tmp/torrent.torrent")

	assert.Nil(t, err)
	rpc.verify(t, "load_start", []interface{}{"/tmp/torrent.torrent"}...)
}

func TestAddTorrentFile(t *testing.T) {
	rt := &rtorrent{}

	rpc := xmlRpcMock{}
	rt.rpc = &rpc
	err := rt.AddFile([]byte("torrent file"))

	assert.Nil(t, err)
	rpc.verify(t, "load_raw_start", []interface{}{[]byte("torrent file")}...)
}

func TestStopTorrent(t *testing.T) {
	rt := &rtorrent{}

	rpc := xmlRpcMock{}
	rt.rpc = &rpc
	err := rt.Stop(api.TorrentDownload{Attrs: map[string]string{"hash": "torrent hash"}})

	assert.Nil(t, err)
	rpc.verify(t, "d.stop", []interface{}{"torrent hash"}...)
}

func TestStartTorrent(t *testing.T) {
	rt := &rtorrent{}

	rpc := xmlRpcMock{}
	rt.rpc = &rpc
	err := rt.Start(api.TorrentDownload{Attrs: map[string]string{"hash": "torrent hash"}})

	assert.Nil(t, err)
	rpc.verify(t, "d.start", []interface{}{"torrent hash"}...)
}

func TestDeleteTorrent(t *testing.T) {
	rt := &rtorrent{}

	rpc := xmlRpcMock{}
	rt.rpc = &rpc
	err := rt.Delete(api.TorrentDownload{Attrs: map[string]string{"hash": "torrent hash"}})

	assert.Nil(t, err)
	rpc.verify(t, "d.erase", []interface{}{"torrent hash"}...)
}

func TestGetTorrents(t *testing.T) {
	rt := &rtorrent{}

	rpc := xmlRpcMock{response: []interface{}{
		[]interface{}{
			[]interface{}{"file1.iso", "./file1.iso", int64(3917479936), int64(3917479936), int64(1), int64(0), int64(1000), "hash1"},
			[]interface{}{"file2.mkv", "./file2.mkv", int64(5117773331), int64(2558886665), int64(0), int64(1), int64(500), "hash2"},
		}},
	}
	rt.rpc = &rpc
	downloads, err := rt.Torrents()

	assert.Nil(t, err)
	assert.NotNil(t, downloads)
	assert.Equal(t,
		[]api.TorrentDownload{
			{Name: "file1.iso", Path: "./file1.iso", Size: 3917479936, CompletedSize: 3917479936, Completed: true, Stopped: true, Ratio: 1, Attrs: map[string]string{"hash": "hash1"}},
			{Name: "file2.mkv", Path: "./file2.mkv", Size: 5117773331, CompletedSize: 2558886665, Completed: false, Stopped: false, Ratio: 0.5, Attrs: map[string]string{"hash": "hash2"}},
		},
		downloads)
	rpc.verify(t, "d.multicall", []interface{}{"main", "d.name=", "d.base_path=", "d.size_bytes=", "d.completed_bytes=", "d.complete=", "d.state=", "d.ratio=", "d.hash="})
}

type xmlRpcMock struct {
	method   string
	args     []interface{}
	response []interface{}
	err      error
}

func (x *xmlRpcMock) Send(method string, args ...interface{}) (params []interface{}, err error) {
	x.method = method
	x.args = args
	params = x.response
	err = x.err
	return
}

func (x *xmlRpcMock) verify(t *testing.T, expectedMethod string, expectedArgs ...interface{}) {
	assert.Equal(t, expectedMethod, x.method)
	assert.Equal(t, expectedArgs, x.args)
}
