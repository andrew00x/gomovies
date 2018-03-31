package player

import (
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/omxcontrol"
	"time"
)

type Factory func(conf *config.Config) (Player, error)

var factory Factory

func Create(conf *config.Config) (Player, error) {
	p, err := factory(conf)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type Player interface {
	AudioTracks() ([]omxcontrol.Stream, error)
	Mute() error
	NextAudioTrack() error
	NextSubtitles() error
	Pause() error
	Play() error
	PlayMovie(path string) error
	PlayPause() error
	PreviousAudioTrack() error
	PreviousSubtitles() error
	ReplayCurrent() error
	Seek(offset time.Duration) error
	SelectAudio(index int) error
	SelectSubtitle(index int) error
	SetPosition(position time.Duration) error
	Status() (api.PlayerStatus, error)
	Stop() error
	Subtitles() ([]omxcontrol.Stream, error)
	ToggleSubtitles() error
	Unmute() error
	Volume() (float64, error)
	VolumeDown() error
	VolumeUp() error
}
