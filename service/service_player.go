package service

import (
	"time"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/gomovies/player"
	"github.com/andrew00x/omxcontrol"
)

type PlayerService struct {
	p player.Player
}

func CreatePlayerService(conf *config.Config) (*PlayerService, error) {
	plr, err := player.Create(conf)
	if err != nil {
		return nil, err
	}
	return createPlayerService(plr), nil
}

func createPlayerService(plr player.Player) *PlayerService {
	return &PlayerService{plr}
}

func (srv *PlayerService) AudioTracks() ([]omxcontrol.Stream, error) {
	return srv.p.AudioTracks()
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

func (srv *PlayerService) PlayMovie(path string) (status api.PlayerStatus, err error) {
	err = srv.p.PlayMovie(path)
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

func (srv *PlayerService) Seek(offset int) (status api.PlayerStatus, err error) {
	err = srv.p.Seek(time.Duration(offset) * time.Second)
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

func (srv *PlayerService) SetPosition(position int) (status api.PlayerStatus, err error) {
	err = srv.p.SetPosition(time.Duration(position) * time.Second)
	if err == nil {
		status, err = srv.p.Status()
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
