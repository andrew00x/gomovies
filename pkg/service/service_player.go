package service

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/gomovies/pkg/player"
	"github.com/andrew00x/omxcontrol"
)

type PlayerService struct {
	player player.Player
	queue  *PlayQueue
}

func CreatePlayerService(conf *config.Config) (*PlayerService, error) {
	p, err := player.Create(conf)
	if err != nil {
		return nil, err
	}
	q := PlayQueue{}
	l := playListener{queue: &q, player: p}
	p.AddListener(&l)
	return createPlayerService(p, &q), nil
}

func createPlayerService(p player.Player, q *PlayQueue) *PlayerService {
	return &PlayerService{player: p, queue: q}
}

type playListener struct {
	queue  *PlayQueue
	player player.Player
}

func (l *playListener) StartPlay(path string) {
}

func (l *playListener) StopPlay(path string) {
	next := l.queue.Pop()
	if next != "" {
		err := l.player.PlayMovie(next)
		if err != nil {
			log.WithFields(log.Fields{"file": next, "err": err}).Error("Unable start play")
		}
	}
}

type PlayQueue struct {
	lock sync.Mutex
	arr  []string
}

func (q *PlayQueue) Enqueue(path string) {
	q.lock.Lock()
	q.arr = append(q.arr, path)
	q.lock.Unlock()
}

func (q *PlayQueue) Dequeue(i int) {
	q.lock.Lock()
	if i < len(q.arr) {
		q.arr = append(q.arr[:i], q.arr[i+1:]...)
	}
	q.lock.Unlock()
}

func (q *PlayQueue) Pop() (path string) {
	q.lock.Lock()
	if len(q.arr) > 0 {
		path, q.arr = q.arr[0], q.arr[1:]
	}
	q.lock.Unlock()
	return
}

func (q *PlayQueue) Empty() (r bool) {
	q.lock.Lock()
	r = len(q.arr) == 0
	q.lock.Unlock()
	return
}

func (q *PlayQueue) All() (all []string) {
	q.lock.Lock()
	all = append([]string{}, q.arr...)
	q.lock.Unlock()
	return
}

func (srv *PlayerService) AudioTracks() ([]omxcontrol.Stream, error) {
	return srv.player.AudioTracks()
}

func (srv *PlayerService) Enqueue(path string) (queue []string, err error) {
	s, _ := srv.player.Status()
	if s.Playing == "" {
		err = srv.player.PlayMovie(path)
	} else {
		srv.queue.Enqueue(path)
		queue = srv.queue.All()
	}
	return
}

func (srv *PlayerService) Dequeue(i int) (queue []string) {
	srv.queue.Dequeue(i)
	queue = srv.queue.All()
	return
}

func (srv *PlayerService) Queue() []string {
	return srv.queue.All()
}

func (srv *PlayerService) NextAudioTrack() (audios []omxcontrol.Stream, err error) {
	err = srv.player.NextAudioTrack()
	if err == nil {
		audios, err = srv.player.AudioTracks()
	}
	return
}

func (srv *PlayerService) NextSubtitles() (subtitles []omxcontrol.Stream, err error) {
	err = srv.player.NextSubtitles()
	if err == nil {
		subtitles, err = srv.player.Subtitles()
	}
	return
}

func (srv *PlayerService) Pause() (status api.PlayerStatus, err error) {
	err = srv.player.Pause()
	if err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) Play() (status api.PlayerStatus, err error) {
	err = srv.player.Play()
	if err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) PlayMovie(path string) (status api.PlayerStatus, err error) {
	err = srv.player.PlayMovie(path)
	if err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) PlayPause() (status api.PlayerStatus, err error) {
	err = srv.player.PlayPause()
	if err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) PreviousAudioTrack() (audios []omxcontrol.Stream, err error) {
	err = srv.player.PreviousAudioTrack()
	if err == nil {
		audios, err = srv.player.AudioTracks()
	}
	return
}

func (srv *PlayerService) PreviousSubtitles() (subtitles []omxcontrol.Stream, err error) {
	err = srv.player.PreviousSubtitles()
	if err == nil {
		subtitles, err = srv.player.Subtitles()
	}
	return
}

func (srv *PlayerService) ReplayCurrent() (status api.PlayerStatus, err error) {
	err = srv.player.ReplayCurrent()
	if err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) Seek(offset int) (status api.PlayerStatus, err error) {
	err = srv.player.Seek(time.Duration(offset) * time.Second)
	if err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) SelectAudio(index int) (audios []omxcontrol.Stream, err error) {
	err = srv.player.SelectAudio(index)
	if err == nil {
		audios, err = srv.player.AudioTracks()
	}
	return
}

func (srv *PlayerService) SelectSubtitle(index int) (subtitles []omxcontrol.Stream, err error) {
	err = srv.player.SelectSubtitle(index)
	if err == nil {
		subtitles, err = srv.player.Subtitles()
	}
	return
}

func (srv *PlayerService) SetPosition(position int) (status api.PlayerStatus, err error) {
	err = srv.player.SetPosition(time.Duration(position) * time.Second)
	if err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) Status() (api.PlayerStatus, error) {
	return srv.player.Status()
}

func (srv *PlayerService) Stop() (status api.PlayerStatus, err error) {
	if err = srv.player.Stop(); err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) Subtitles() ([]omxcontrol.Stream, error) {
	return srv.player.Subtitles()
}

func (srv *PlayerService) ToggleSubtitles() (status api.PlayerStatus, err error) {
	if err = srv.player.ToggleSubtitles(); err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) ToggleMute() (status api.PlayerStatus, err error) {
	if err = srv.player.ToggleMute(); err == nil {
		status, err = srv.player.Status()
	}
	return
}

func (srv *PlayerService) Volume() (float64, error) {
	return srv.player.Volume()
}

func (srv *PlayerService) VolumeDown() (vol float64, err error) {
	err = srv.player.VolumeDown()
	if err == nil {
		vol, err = srv.player.Volume()
	}
	return
}

func (srv *PlayerService) VolumeUp() (vol float64, err error) {
	err = srv.player.VolumeUp()
	if err == nil {
		vol, err = srv.player.Volume()
	}
	return
}
