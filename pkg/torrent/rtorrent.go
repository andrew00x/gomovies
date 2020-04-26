package torrent

import (
	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/xmlrpc/pkg/xmlrpc"
)

type rtorrent struct {
	rpc xmlrpc.Client
}

func init() {
	Factory = createRtorrent
}

func createRtorrent(cfg *config.Config) Torrent {
	return &rtorrent{
		rpc: xmlrpc.CreateSCGIClient(cfg.TorrentRemoteCtrlAddr),
	}
}

func (t *rtorrent) AddFile(b []byte) error {
	_, err := t.rpc.Send("load_raw_start", []interface{}{b}...)
	return err
}

func (t *rtorrent) AddUrl(u string) error {
	_, err := t.rpc.Send("load_start", []interface{}{u}...)
	return err
}

func (t *rtorrent) Torrents() ([]api.TorrentDownload, error) {
	res, err := t.rpc.Send("d.multicall",
		[]interface{}{"main", "d.name=", "d.base_path=", "d.size_bytes=", "d.completed_bytes=", "d.complete=", "d.state=", "d.ratio=", "d.hash=", "d.message="})
	if err != nil {
		return nil, err
	}
	downloads := make([]api.TorrentDownload, 0)
	for _, item := range res[0].([]interface{}) {
		data := item.([]interface{})
		downloads = append(downloads, api.TorrentDownload{
			Name:          data[0].(string),
			Path:          data[1].(string),
			Size:          data[2].(int64),
			CompletedSize: data[3].(int64),
			Completed:     data[4].(int64) > 0,
			Stopped:       data[5].(int64) == 0,
			Ratio:         float32(data[6].(int64)) / 1000,
			Attrs: map[string]string{
				"hash":    data[7].(string),
				"message": data[8].(string),
			},
		})
	}
	return downloads, nil
}

func (t *rtorrent) Start(d api.TorrentDownload) error {
	_, err := t.rpc.Send("d.start", []interface{}{d.Attrs["hash"]}...)
	return err
}

func (t *rtorrent) Stop(d api.TorrentDownload) error {
	_, err := t.rpc.Send("d.stop", []interface{}{d.Attrs["hash"]}...)
	return err
}

func (t *rtorrent) Delete(d api.TorrentDownload) error {
	_, err := t.rpc.Send("d.erase", []interface{}{d.Attrs["hash"]}...)
	return err
}
