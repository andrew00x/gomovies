package player

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/omxcontrol"
)

type OMXPlayer struct {
	process      *os.Process
	control      *omxcontrol.OmxCtrl
	listeners    []PlayListener
	muted        bool
	subtitlesOff bool
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

func (p *OMXPlayer) AudioTracks() (audios []api.Stream, err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		var controlAudios []omxcontrol.Stream
		controlAudios, err = control.AudioTracks()
		if err == nil {
			audios = convertToApiStreams(controlAudios)
		}
	}
	return
}

func (p *OMXPlayer) NextAudioTrack() error {
	return p.action(omxcontrol.ActionNextAudio)
}

func (p *OMXPlayer) NextSubtitle() error {
	return p.action(omxcontrol.ActionNextSubtitle)
}

func (p *OMXPlayer) Pause() (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		err = control.Pause()
	}
	return
}

func (p *OMXPlayer) Play() (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		err = control.Play()
	}
	return
}

func (p *OMXPlayer) PlayMovie(path string) (err error) {
	if stpErr := p.Stop(); stpErr != nil {
		log.WithFields(log.Fields{"err": stpErr}).Error("Error occurred while stopping player")
	}
	err = p.start(path)
	if err == nil {
		var control *omxcontrol.OmxCtrl
		control, err = setupControl()
		if err == nil {
			p.control = control
		} else {
			log.WithFields(log.Fields{"err": err}).Error("Error occurred while setup DBus connection to omxplayer")
			if stpErr := p.Stop(); stpErr != nil {
				log.WithFields(log.Fields{"err": stpErr}).Error("Error occurred while trying to stop player after unsuccessful start")
			}
		}
	}
	if err == nil {
		for _, l := range p.listeners {
			l.StartPlay(path)
		}
		go func() {
			if _, waitErr := p.process.Wait(); waitErr != nil {
				log.WithFields(log.Fields{"err": err}).Error("Player process ended with error")
			}
			for _, l := range p.listeners {
				l.StopPlay(path)
			}
		}()
	}
	return
}

func (p *OMXPlayer) PlayPause() (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		err = control.PlayPause()
	}
	return
}

func (p *OMXPlayer) PreviousAudioTrack() error {
	return p.action(omxcontrol.ActionPreviousAudio)
}

func (p *OMXPlayer) PreviousSubtitle() error {
	return p.action(omxcontrol.ActionPreviousSubtitle)
}

func (p *OMXPlayer) ReplayCurrent() (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		err = control.SetPosition(0)
	}
	return
}

func (p *OMXPlayer) Seek(offset time.Duration) (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		err = control.Seek(offset)
	}
	return
}

func (p *OMXPlayer) SelectAudio(index int) (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		var ok bool
		if ok, err = control.SelectAudio(index); ok {
		} else {
			err = fmt.Errorf("audio track %d was not selected", index)
		}
	}
	return
}

func (p *OMXPlayer) SelectSubtitle(index int) (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		var ok bool
		if ok, err = control.SelectSubtitle(index); ok {
		} else {
			err = fmt.Errorf("subtitle %d was not selected", index)
		}
	}
	return
}

func (p *OMXPlayer) SetPosition(position time.Duration) (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		err = control.SetPosition(position)
	}
	return
}

func (p *OMXPlayer) Status() (status api.PlayerStatus, err error) {
	status.Stopped = true
	control := p.control
	if control != nil {
		if ready, e := control.CanControl(); ready && e == nil {
			var file string
			var position, duration time.Duration
			var pbs omxcontrol.Status
			var audios []omxcontrol.Stream
			var subs []omxcontrol.Stream
			file, err = control.Playing()
			if err != nil {
				return
			}
			position, err = control.Position()
			if err != nil {
				return
			}
			duration, err = control.Duration()
			if err != nil {
				return
			}
			pbs, err = control.PlaybackStatus()
			if err != nil {
				return
			}
			audios, err = control.AudioTracks()
			if err != nil {
				return
			}
			subs, err = control.Subtitles()
			if err == nil {
				status.File = file
				status.Position = int(position / time.Second)
				status.Duration = int(duration / time.Second)
				status.Paused = pbs == omxcontrol.Paused
				status.Muted = p.muted
				status.SubtitlesOff = p.subtitlesOff
				status.ActiveAudioTrack = findActive(audios)
				status.ActiveSubtitle = findActive(subs)
				status.Stopped = false
			}
		}
	}
	return
}

func (p *OMXPlayer) Stop() error {
	return p.quit()
}

func (p *OMXPlayer) Subtitles() (subtitles []api.Stream, err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		var controlSubtitles []omxcontrol.Stream
		controlSubtitles, err = control.Subtitles()
		if err == nil {
			subtitles = convertToApiStreams(controlSubtitles)
		}
	}
	return
}

func (p *OMXPlayer) ToggleMute() (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		if p.muted {
			err = control.Unmute()
		} else {
			err = control.Mute()
		}
	}
	if err == nil {
		p.muted = !p.muted
	}
	return
}

func (p *OMXPlayer) ToggleSubtitles() (err error) {
	if err = p.action(omxcontrol.ActionToggleSubtitle); err == nil {
		p.subtitlesOff = !p.subtitlesOff
	}
	return
}

func (p *OMXPlayer) Volume() (vol float64, err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		vol, err = control.Volume()
	}
	return
}

func (p *OMXPlayer) VolumeDown() error {
	return p.action(omxcontrol.ActionDecreaseVolume)
}

func (p *OMXPlayer) VolumeUp() error {
	return p.action(omxcontrol.ActionIncreaseVolume)
}

func (p *OMXPlayer) action(actionCode omxcontrol.KeyboardAction) (err error) {
	var control *omxcontrol.OmxCtrl
	if control, err = p.mustHaveControl(); err == nil {
		err = control.Action(actionCode)
	}
	return
}

func (p *OMXPlayer) quit() (err error) {
	process := p.process
	if process != nil {
		log.WithFields(log.Fields{"PID": process.Pid}).Info("kill omxplayer")
		var pgid int
		if pgid, err = syscall.Getpgid(process.Pid); err != nil {
			return
		}
		if err = syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
			return
		}
		_, _ = process.Wait()
		p.process = nil
		p.control = nil
		p.muted = false
		p.subtitlesOff = false
	}
	return
}

func setupControl() (control *omxcontrol.OmxCtrl, err error) {
	attempts := 10
	retryDelay := time.Duration(2) * time.Second
	var ready bool
	for i := 1; ; i++ {
		time.Sleep(retryDelay)
		control, err = omxcontrol.Create()
		ready, err = control.CanControl()
		if err == nil && ready {
			log.WithFields(log.Fields{"attempts": i}).Info("Setup omxplayer control")
			return
		}
		if i > attempts {
			break
		}
	}
	err = fmt.Errorf("unable setup omxplayer control after %d attempts, last error: %v", attempts, err)
	return
}

func (p *OMXPlayer) start(path string) (err error) {
	cmd := exec.Command("/usr/bin/omxplayer", "-b", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err == nil {
		p.process = cmd.Process
		log.WithFields(log.Fields{"PID": p.process.Pid, "file": path}).Info("Started omxplayer")
	}
	return
}

func (p *OMXPlayer) mustHaveControl() (control *omxcontrol.OmxCtrl, err error) {
	control = p.control
	if control == nil {
		err = controlNotSetup
	}
	return
}

func findActive(streams []omxcontrol.Stream) int {
	for _, stream := range streams {
		if stream.Active {
			return stream.Index
		}
	}
	return -1
}

func convertToApiStreams(streams []omxcontrol.Stream) []api.Stream {
	var apiStreams []api.Stream
	for _, stream := range streams {
		apiStreams = append(apiStreams, api.Stream{Index: stream.Index, Name: stream.Name, Language: stream.Language, Codec: stream.Codec, Active: stream.Active})
	}
	return apiStreams
}
