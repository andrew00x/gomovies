package player

import (
	"time"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/omxcontrol"
)

type Factory func(conf *config.Config) (Player, error)

var playerFactory Factory

func Create(conf *config.Config) (Player, error) {
	return playerFactory(conf)
}

type Player interface {
	AudioTracks() ([]omxcontrol.Stream, error)
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
	ToggleMute() error
	ToggleSubtitles() error
	Volume() (float64, error)
	VolumeDown() error
	VolumeUp() error
	Observable
}

type Observable interface {
	AddListener(l PlayListener)
}

type StartPlayListener interface {
	StartPlay(path string)
}

type StopPlayListener interface {
	StopPlay(path string)
}

type PlayListener interface {
	StartPlayListener
	StopPlayListener
}
