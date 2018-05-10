package player

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/omxcontrol"
)

type OMXPlayer struct {
	process   *os.Process
	control   *omxcontrol.OmxCtrl
	listeners []PlayListener
}

func init() {
	playerFactory = func(conf *config.Config) (Player, error) {
		return &OMXPlayer{}, nil
	}
}

var controlNotSetup = errors.New("omxplayer does not play anything at the moment or control is not setup")

func (p *OMXPlayer) AddListener(l PlayListener) {
	p.listeners = append(p.listeners, l)
}

func (p *OMXPlayer) AudioTracks() (audios []omxcontrol.Stream, err error) {
	if p.control == nil {
		err = controlNotSetup
	} else {
		audios, err = p.control.AudioTracks()
	}
	return
}

func (p *OMXPlayer) Mute() (err error) {
	if p.control == nil {
		err = controlNotSetup
	} else {
		err = p.control.Mute()
	}
	return
}

func (p *OMXPlayer) NextAudioTrack() error {
	return p.action(omxcontrol.ActionNextAudio)
}

func (p *OMXPlayer) NextSubtitles() error {
	return p.action(omxcontrol.ActionNextSubtitle)
}

func (p *OMXPlayer) Pause() error {
	if p.control == nil {
		return controlNotSetup
	}
	return p.control.Pause()
}

func (p *OMXPlayer) Play() error {
	if p.control == nil {
		return controlNotSetup
	}
	return p.control.Play()
}

func (p *OMXPlayer) PlayMovie(path string) (err error) {
	p.Stop()
	err = p.start(path)
	if err == nil {
		var control *omxcontrol.OmxCtrl
		control, err = setupControl()
		if err == nil {
			p.control = control
		} else {
			p.quit()
		}
	}
	if err == nil {
		for _, l := range p.listeners {
			l.StartPlay(path)
		}
		go func() {
			p.process.Wait()
			for _, l := range p.listeners {
				l.StopPlay(path)
			}
		}()
	}
	return
}

func (p *OMXPlayer) PlayPause() error {
	if p.control == nil {
		return controlNotSetup
	}
	return p.control.PlayPause()
}

func (p *OMXPlayer) PreviousAudioTrack() error {
	return p.action(omxcontrol.ActionPreviousAudio)
}

func (p *OMXPlayer) PreviousSubtitles() error {
	return p.action(omxcontrol.ActionPreviousSubtitle)
}

func (p *OMXPlayer) ReplayCurrent() (err error) {
	if p.control == nil {
		err = controlNotSetup
	} else {
		err = p.control.SetPosition(0)
	}
	return
}

func (p *OMXPlayer) Seek(offset time.Duration) error {
	if p.control == nil {
		return controlNotSetup
	}
	return p.control.Seek(offset)
}

func (p *OMXPlayer) SelectAudio(index int) (err error) {
	if p.control == nil {
		err = controlNotSetup
	} else {
		var ok bool
		if ok, err = p.control.SelectAudio(index); !ok {
			err = errors.New(fmt.Sprintf("audio track %d was not selected", index))
		}
	}
	return
}

func (p *OMXPlayer) SelectSubtitle(index int) (err error) {
	if p.control == nil {
		err = controlNotSetup
	} else {
		var ok bool
		if ok, err = p.control.SelectSubtitle(index); !ok {
			err = errors.New(fmt.Sprintf("subtitle %d was not selected", index))
		}
	}
	return
}

func (p *OMXPlayer) SetPosition(position time.Duration) error {
	if p.control == nil {
		return controlNotSetup
	}
	return p.control.SetPosition(position)
}

func (p *OMXPlayer) Status() (status api.PlayerStatus, err error) {
	if p.control != nil {
		var playing string
		var position, duration time.Duration
		var pbs omxcontrol.Status
		playing, err = p.control.Playing()
		position, err = p.control.Position()
		duration, err = p.control.Duration()
		pbs, err = p.control.PlaybackStatus()
		if err == nil {
			status.Playing = playing
			status.Position = int(position / time.Second)
			status.Duration = int(duration / time.Second)
			status.Paused = pbs == omxcontrol.Paused
		}
	}
	return
}

func (p *OMXPlayer) Stop() error {
	return p.quit()
}

func (p *OMXPlayer) Subtitles() (subtitles []omxcontrol.Stream, err error) {
	if p.control == nil {
		err = controlNotSetup
	} else {
		subtitles, err = p.control.Subtitles()
	}
	return
}

func (p *OMXPlayer) Unmute() (err error) {
	if p.control == nil {
		err = controlNotSetup
	} else {
		err = p.control.Unmute()
	}
	return
}

func (p *OMXPlayer) ToggleSubtitles() error {
	return p.action(omxcontrol.ActionToggleSubtitle)
}

func (p *OMXPlayer) Volume() (vol float64, err error) {
	if p.control == nil {
		err = controlNotSetup
	} else {
		vol, err = p.control.Volume()
	}
	return
}

func (p *OMXPlayer) VolumeDown() error {
	return p.action(omxcontrol.ActionDecreaseVolume)
}

func (p *OMXPlayer) VolumeUp() error {
	return p.action(omxcontrol.ActionIncreaseVolume)
}

func (p *OMXPlayer) action(actionCode omxcontrol.KeyboardAction) error {
	if p.control == nil {
		return controlNotSetup
	}
	return p.control.Action(actionCode)
}

func (p *OMXPlayer) quit() (err error) {
	process := p.process
	if process != nil {
		log.Printf("kill omxplayer, pid: (%d)\n", process.Pid)
		pgid, err := syscall.Getpgid(process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		}
		process.Wait()
		p.process = nil
		p.control = nil
	}
	return
}

func setupControl() (control *omxcontrol.OmxCtrl, err error) {
	attempts := 50
	retryDelay := time.Duration(100) * time.Millisecond
	var ready bool
	for i := 1; ; i++ {
		time.Sleep(retryDelay)
		control, err = omxcontrol.Create()
		ready, err = control.CanControl()
		if err == nil && ready {
			log.Printf("setup omxplayer control after %d attempts\n", i)
			return
		}
		if i > attempts {
			break
		}
	}
	err = errors.New(fmt.Sprintf("unable setup omxplayer control after %d attempts, last error: %v", attempts, err))
	return
}

func (p *OMXPlayer) start(path string) (err error) {
	cmd := exec.Command("/usr/bin/omxplayer", "-b", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err == nil {
		p.process = cmd.Process
		log.Printf("started omxplayer, pid: (%d); playing: %s\n", p.process.Pid, path)
	}
	return
}

/*
func (p *OMXPlayer) isRunning() (alive bool) {
	if err := p.process.Signal(syscall.Signal(0)); err == nil {
		alive = true
	}
	return
}
*/