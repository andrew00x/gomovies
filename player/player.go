package player

import (
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
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
	Status() (*api.PlayerStatus, error)

	Play(path string) error
	Enqueue(path string) error

	Stop() error
	PlayPause() error
	ReplayCurrent() error
	Forward30s() error
	Rewind30s() error
	Forward10m() error
	Rewind10m() error
	VolumeUp() error
	VolumeDown() error
	NextAudioTrack() error
	PreviousAudioTrack() error
	NextSubtitles() error
	PreviousSubtitles() error
	ToggleSubtitles() error

	Playlist() (*api.Playlist, error)
	NextInPlaylist() error
	PreviousInPlaylist() error
	DeleteInPlaylist(pos int) error
	PlayInPlaylist(pos int) error
}
