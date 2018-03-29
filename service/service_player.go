package service

import (
	"errors"
	"fmt"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/catalog"
	"github.com/andrew00x/gomovies/player"
	"github.com/andrew00x/omxcontrol"
)

type PlayerService struct {
	p   player.Player
	ctl *catalog.Catalog
}

func CreatePlayerService(plr player.Player, ctl *catalog.Catalog) *PlayerService {
	return &PlayerService{plr, ctl}
}

func (srv *PlayerService) AudioTracks() ([]omxcontrol.Stream, error) {
	return srv.p.AudioTracks()
}

func (srv *PlayerService) Forward10m() (status api.PlayerStatus, err error) {
	err = srv.p.Forward10m()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) Forward30s() (status api.PlayerStatus, err error) {
	err = srv.p.Forward30s()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) Mute() error {
	return srv.p.Mute()
}

func (srv *PlayerService) NextAudioTrack() (audios []omxcontrol.Stream, err error) {
	err = srv.p.NextAudioTrack()
	if err == nil {
		audios, err = srv.p.AudioTracks()
	}
	return
}

func (srv *PlayerService) NextSubtitles() (subtitles []omxcontrol.Stream, err error) {
	err = srv.p.NextSubtitles()
	if err == nil {
		subtitles, err = srv.p.Subtitles()
	}
	return
}

func (srv *PlayerService) Pause() (status api.PlayerStatus, err error) {
	err = srv.p.Pause()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) Play() (status api.PlayerStatus, err error) {
	err = srv.p.Play()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) PlayMovie(id int) (status api.PlayerStatus, err error) {
	m := srv.ctl.Get(id)
	if m == nil {
		err = errors.New(fmt.Sprintf("invalid movie id: %d", id))
		return
	}
	err = srv.p.PlayMovie(m.Path)
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) PlayPause() (status api.PlayerStatus, err error) {
	err = srv.p.PlayPause()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) PreviousAudioTrack() (audios []omxcontrol.Stream, err error) {
	err = srv.p.PreviousAudioTrack()
	if err == nil {
		audios, err = srv.p.AudioTracks()
	}
	return
}

func (srv *PlayerService) PreviousSubtitles() (subtitles []omxcontrol.Stream, err error) {
	err = srv.p.PreviousSubtitles()
	if err == nil {
		subtitles, err = srv.p.Subtitles()
	}
	return
}

func (srv *PlayerService) ReplayCurrent() (status api.PlayerStatus, err error) {
	err = srv.p.ReplayCurrent()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) Rewind10m() (status api.PlayerStatus, err error) {
	err = srv.p.Rewind10m()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) Rewind30s() (status api.PlayerStatus, err error) {
	err = srv.p.Rewind30s()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) SelectAudio(index int) (audios []omxcontrol.Stream, err error) {
	err = srv.p.SelectAudio(index)
	if err == nil {
		audios, err = srv.p.AudioTracks()
	}
	return
}

func (srv *PlayerService) SelectSubtitle(index int) (subtitles []omxcontrol.Stream, err error) {
	err = srv.p.SelectSubtitle(index)
	if err == nil {
		subtitles, err = srv.p.Subtitles()
	}
	return
}

func (srv *PlayerService) Status() (api.PlayerStatus, error) {
	return srv.p.Status()
}

func (srv *PlayerService) Stop() (status api.PlayerStatus, err error) {
	err = srv.p.Stop()
	if err == nil {
		status, err = srv.p.Status()
	}
	return
}

func (srv *PlayerService) Subtitles() ([]omxcontrol.Stream, error) {
	return srv.p.Subtitles()
}

func (srv *PlayerService) ToggleSubtitles() error {
	return srv.p.ToggleSubtitles()
}

func (srv *PlayerService) Unmute() error {
	return srv.p.Unmute()
}

func (srv *PlayerService) Volume() (float64, error) {
	return srv.p.Volume()
}

func (srv *PlayerService) VolumeDown() (vol float64, err error) {
	err = srv.p.VolumeDown()
	if err == nil {
		vol, err = srv.p.Volume()
	}
	return
}

func (srv *PlayerService) VolumeUp() (vol float64, err error) {
	err = srv.p.VolumeUp()
	if err == nil {
		vol, err = srv.p.Volume()
	}
	return
}
