package player

import (
	"time"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/omxcontrol"
)

type PlayerFactory func(conf *config.Config) (Player, error)

var playerFactory PlayerFactory

func Create(conf *config.Config) (Player, error) {
	return playerFactory(conf)
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
