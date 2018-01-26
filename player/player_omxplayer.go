package player

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"regexp"
	"time"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/file"
	"github.com/andrew00x/gomovies/config"
)

var omxCtl = "/var/run/omxctl"
var omxStat = "/var/log/omxstat"
var omxPlay = "/var/local/omxplay"

type OMXPlayer struct {
	savePlaybackTime bool
	playbacks        map[string]*playback
}

type playback struct {
	startPos    int
}

func init() {
	factory = func(conf *config.Config) (Player, error) {
		pid, err := pidOf("omxd")
		if err != nil {
			return nil, err
		}
		if pid == 0 {
			return nil, errors.New(fmt.Sprintf("Unable detect 'omxd' status: %v\n", err))
		}
		return &OMXPlayer{savePlaybackTime: conf.SavePlaybackTime == "yes", playbacks: make(map[string]*playback)}, nil
	}
}

func pidOf(cmd string) (int, error) {
	pgrep := exec.Command("pgrep", cmd)
	pid, err := pgrep.Output()
	if err != nil {
		return 0, errors.New("please check that 'omxd' is running")
	}
	s := strings.Fields(string(pid))[0]
	return strconv.Atoi(s)
}

func (p *OMXPlayer) Status() (*api.PlayerStatus, error) {
	st, err := readStatus(omxStat)
	if err != nil {
		return nil, err
	}
	if st.Playing != "" {
		pb, ok := p.playbacks[st.Playing]
		if ok {
			st.Position += pb.startPos
		}
	}
	return st, nil
}

func (p *OMXPlayer) Play(playIt string) error {
	var cmd string
	pb, ok := p.playbacks[playIt]
	if ok {
		cmd = fmt.Sprintf(`O
O -b
O -l
O %s
H %s
`, hoursMinsSecs(pb.startPos), playIt)
	} else {
		cmd = fmt.Sprintf(`O
O -b
H %s
`, playIt)
	}
	return sendCommand(cmd)
}

func hoursMinsSecs(t int) string {
	h := t / 3600
	m := (t % 3600) / 60
	s := t % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (p *OMXPlayer) Enqueue(enqueueIt string) error {
	cmd := fmt.Sprintf("A %s\n", enqueueIt)
	return sendCommand(cmd)
}

func (p *OMXPlayer) Stop() error {
	if p.savePlaybackTime {
		st, err := p.Status()
		if err != nil {
			log.Printf("Unable read current status of player: %v\nSkip saving playback time", err)
		} else {
			pb, ok := p.playbacks[st.Playing]
			if !ok {
				pb = &playback{}
				p.playbacks[st.Playing] = pb
			}
			pb.startPos = st.Position
		}
	}
	return sendCommand("P\n")
}

func (p *OMXPlayer) PlayPause() error {
	return sendCommand("p\n")
}

func (p *OMXPlayer) ReplayCurrent() error {
	st, err := p.Status()
	if err != nil {
		return err
	}
	if st.Playing != "" {
		delete(p.playbacks, st.Playing)
		cmd := fmt.Sprintf(`O
O -b
O -l
O 00:00:00
H %s
`, st.Playing)
		return sendCommand(cmd)
	}
	return nil
}

func (p *OMXPlayer) Forward30s() error {
	return sendCommand("f\n")
}

func (p *OMXPlayer) Rewind30s() error {
	return sendCommand("r\n")
}

func (p *OMXPlayer) Forward10m() error {
	return sendCommand("F\n")
}

func (p *OMXPlayer) Rewind10m() error {
	return sendCommand("R\n")
}

func (p *OMXPlayer) VolumeUp() error {
	return sendCommand("+\n")
}

func (p *OMXPlayer) VolumeDown() error {
	return sendCommand("-\n")
}

func (p *OMXPlayer) NextAudioTrack() error {
	return sendCommand("k\n")
}

func (p *OMXPlayer) PreviousAudioTrack() error {
	return sendCommand("K\n")
}

func (p *OMXPlayer) NextSubtitles() error {
	return sendCommand("m\n")
}

func (p *OMXPlayer) PreviousSubtitles() error {
	return sendCommand("M\n")
}

func (p *OMXPlayer) ToggleSubtitles() error {
	return sendCommand("s\n")
}

func (p *OMXPlayer) Playlist() (*api.Playlist, error) {
	playlist, err := readPlaylist(omxPlay)
	if err != nil && os.IsNotExist(err) {
		return &api.Playlist{Items: []string{}}, nil
	} else if err != nil {
		return nil, err
	}
	return playlist, nil
}

func (p *OMXPlayer) NextInPlaylist() error {
	return sendCommand("n\n")
}

func (p *OMXPlayer) PreviousInPlaylist() error {
	return sendCommand("N\n")
}

func (p *OMXPlayer) DeleteInPlaylist(pos int) error {
	cmd := fmt.Sprintf("x %d\n", pos)
	return sendCommand(cmd)
}

func (p *OMXPlayer) PlayInPlaylist(pos int) error {
	cmd := fmt.Sprintf("g %d\n", pos)
	return sendCommand(cmd)
}

func sendCommand(cmd string) error {
	f, err := os.OpenFile(omxCtl, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.WriteString(cmd)
	if err == nil {
		err = w.Flush()
	}
	return err
}

func readStatus(omxstat string) (*api.PlayerStatus, error) {
	var logFile string
	status := &api.PlayerStatus{}
	err := file.ReadLines(omxstat, func(line string) bool {
		// Format: timestamp state [dt logfile file]: '%d %s\n' or '%d %s %d %s %s\n`
		var p int
		var start, st, play, playing string
		p = strings.IndexRune(line, ' ')
		if p > 0 {
			start = line[0:p]
			line = line[p+1:]
		}
		p = strings.IndexRune(line, ' ')
		if p > 0 {
			st = line[0:p]
		} else {
			st = line
		}
		if st != "" && st != "Stopped" {
			line = line[p+1:]
			p = strings.IndexRune(line, ' ')
			play = line[0:p]
			line = line[p+1:]
			p = strings.IndexRune(line, ' ')
			logFile = line[0:p]
			line = line[p+1:]
			playing = line
		}
		status.Paused = st == "Paused"
		status.Playing = playing
		if st == "Playing" {
			startTime, _ := strconv.ParseInt(start, 10, 64)
			playTime, _ := strconv.ParseInt(play, 10, 64)
			status.Position = int(playTime) + int(time.Now().Unix()-startTime)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	if logFile != "" {
		err := addVideoFileInfo(logFile, status)
		if err != nil {
			return nil, err
		}
	}
	return status, nil
}

func addVideoFileInfo(videoLogFile string, status *api.PlayerStatus) (err error) {
	streamRegex := regexp.MustCompile(`Stream #\d+:(\d+)(\(\w+\))?: (Audio|Video|Subtitle): (.*)?`)
	err = file.ReadLines(videoLogFile, func(line string) bool {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Duration:") {
			ds := line[10:18]
			var h, m, s int
			fmt.Sscanf(ds, "%d:%d:%d", &h, &m, &s)
			d := h * 3600
			d += m * 60
			d += s
			status.Duration = d
		} else if strings.HasPrefix(line, "Stream") {
			streamLine := streamRegex.FindStringSubmatch(line)
			if len(streamLine) == 5 {
				num64, _ := strconv.ParseInt(streamLine[1], 10, 64)
				num := int(num64)
				lang, sType, details := streamLine[2], streamLine[3], streamLine[4]
				if sType == "Audio" || sType == "Subtitle" {
					stream := api.Stream{Num: num,
						Lang: strings.Trim(lang, "()"),
						Type: sType,
						Default: strings.HasSuffix(details, "(default)")}
					status.Streams = append(status.Streams, &stream)
				}
			}
		}
		return true
	})
	if err == nil {
		for _, t := range []string{"Audio", "Subtitle"} {
			ofType := filter(status.Streams, func(s api.Stream) bool {
				return s.Type == t
			})
			if len(ofType) == 1 {
				ofType[0].Default = true
			}
		}
	}
	return
}

func filter(streams []*api.Stream, predicate func(api.Stream) bool) (res []*api.Stream) {
	for _, s := range streams {
		if predicate(*s) {
			res = append(res, s)
		}
	}
	return
}

func readPlaylist(omxplay string) (*api.Playlist, error) {
	items := make([]string, 0, 10)
	err := file.ReadLines(omxplay, func(line string) bool {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return &api.Playlist{Items: items}, err
}
