package service

import (
	"errors"
	"fmt"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/catalog"
	"github.com/andrew00x/gomovies/player"
)

type PlayerService struct {
	p   player.Player
	ctl *catalog.Catalog
}

func CreatePlayerService(plr player.Player, ctl *catalog.Catalog) *PlayerService {
	return &PlayerService{plr, ctl}
}

func (srv *PlayerService) Status() (*api.PlayerStatus, error) {
	return srv.p.Status()
}

func (srv *PlayerService) Play(id int) error {
	m := srv.ctl.Get(id)
	if m == nil {
		return errors.New(fmt.Sprintf("invalid movie id: %d", id))
	}
	return srv.p.Play(m.Path)
}

func (srv *PlayerService) Enqueue(id int) error {
	m := srv.ctl.Get(id)
	if m == nil {
		return errors.New(fmt.Sprintf("invalid movie id: %d", id))
	}
	return srv.p.Enqueue(m.Path)
}

func (srv *PlayerService) PlayPause() error {
	return srv.p.PlayPause()
}

func (srv *PlayerService) StopPlay() error {
	return srv.p.Stop()
}

func (srv *PlayerService) ReplayCurrent() error {
	return srv.p.ReplayCurrent()
}

func (srv *PlayerService) Forward30s() error {
	return srv.p.Forward30s()
}

func (srv *PlayerService) Rewind30s() error {
	return srv.p.Rewind30s()
}

func (srv *PlayerService) Forward10m() error {
	return srv.p.Forward10m()
}

func (srv *PlayerService) Rewind10m() error {
	return srv.p.Rewind10m()
}

func (srv *PlayerService) VolumeUp() error {
	return srv.p.VolumeUp()
}

func (srv *PlayerService) VolumeDown() error {
	return srv.p.VolumeDown()
}

func (srv *PlayerService) NextAudioTrack() error {
	return srv.p.NextAudioTrack()
}

func (srv *PlayerService) PreviousAudioTrack() error {
	return srv.p.PreviousAudioTrack()
}

func (srv *PlayerService) NextSubtitles() error {
	return srv.p.NextSubtitles()
}

func (srv *PlayerService) PreviousSubtitles() error {
	return srv.p.PreviousSubtitles()
}

func (srv *PlayerService) ToggleSubtitles() error {
	return srv.p.ToggleSubtitles()
}

func (srv *PlayerService) Playlist() (*api.Playlist, error) {
	return srv.p.Playlist()
}

func (srv *PlayerService) NextInPlaylist() error {
	return srv.p.NextInPlaylist()
}

func (srv *PlayerService) PreviousInPlaylist() error {
	return srv.p.PreviousInPlaylist()
}

func (srv *PlayerService) DeleteInPlaylist(pos int) error {
	return srv.p.DeleteInPlaylist(pos)
}

func (srv *PlayerService) PlayInPlaylist(pos int) error {
	return srv.p.PlayInPlaylist(pos)
}
