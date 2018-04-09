package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"net/http"
	"os/signal"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/gomovies/service"
	"github.com/andrew00x/gomovies/catalog"
	"github.com/andrew00x/gomovies/player"
	"github.com/andrew00x/omxcontrol"
)

var conf *config.Config
var catalogService *service.CatalogService
var playerService *service.PlayerService

func main() {
	var err error
	conf, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not read configuration: %v", err)
	}

	ctl, err := catalog.CreateCatalog(conf)
	if err != nil {
		log.Fatalf("Could not create catalog: %v", err)
	}
	catalogService = service.CreateCatalogService(ctl)

	var plr player.Player
	plr, err = player.Create(conf)
	if err != nil {
		log.Fatalf("Could not create player: %v", err)
	}
	playerService = service.CreatePlayerService(plr, ctl)

	web := http.Server{Addr: fmt.Sprintf(":%d", conf.WebPort), Handler: http.DefaultServeMux}
	log.Printf("Starting on port: %d\n", conf.WebPort)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	go func() {
		<-quit
		if err := catalogService.Stop(); err != nil {
			log.Printf("Unable save catalog file: %v\n", err)
		} else {
			log.Println("Save catalog file")
		}
		if err := web.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not shutdown: %v", err)
		}
	}()

	webClient()
	http.HandleFunc("/api/list", allMovies)
	http.HandleFunc("/api/play", playMovie)
	http.HandleFunc("/api/search", searchMovies)
	http.HandleFunc("/api/refresh", refresh)
	http.HandleFunc("/api/player/audios", audios)
	http.HandleFunc("/api/player/mute", mute)
	http.HandleFunc("/api/player/nextaudiotrack", nextAudioTrack)
	http.HandleFunc("/api/player/nextsubtitles", nextSubtitles)
	http.HandleFunc("/api/player/pause", pause)
	http.HandleFunc("/api/player/play", play)
	http.HandleFunc("/api/player/playpause", playPause)
	http.HandleFunc("/api/player/previousaudiotrack", previousAudioTrack)
	http.HandleFunc("/api/player/previoussubtitles", previousSubtitles)
	http.HandleFunc("/api/player/replay", replayCurrent)
	http.HandleFunc("/api/player/seek", seek)
	http.HandleFunc("/api/player/audio", selectAudio)
	http.HandleFunc("/api/player/subtitle", selectSubtitle)
	http.HandleFunc("/api/player/position", setPosition)
	http.HandleFunc("/api/player/status", status)
	http.HandleFunc("/api/player/stop", stop)
	http.HandleFunc("/api/player/subtitles", subtitles)
	http.HandleFunc("/api/player/unmute", unmute)
	http.HandleFunc("/api/player/togglesubtitles", toggleSubtitles)
	http.HandleFunc("/api/player/volume", volume)
	http.HandleFunc("/api/player/volumedown", volumeDown)
	http.HandleFunc("/api/player/volumeup", volumeUp)

	if err = web.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Could not start http listener: %v\n", err)
	}
}

func webClient() {
	webDir := conf.WebDir
	if webDir != "" {
		log.Printf("Starting web client from directory: %s\n", webDir)
		fs := http.FileServer(http.Dir(webDir))
		http.Handle("/", fs)
	}
}

func allMovies(w http.ResponseWriter, _ *http.Request) {
	writeJsonResponse(catalogService.All(), nil, w)
}

func audios(w http.ResponseWriter, _ *http.Request) {
	audios, err := playerService.AudioTracks()
	writeJsonResponse(audios, err, w)
}

func mute(w http.ResponseWriter, _ *http.Request) {
	err := playerService.Mute()
	writeJsonResponse(nil, err, w)
}

func nextAudioTrack(w http.ResponseWriter, _ *http.Request) {
	audios, err := playerService.NextAudioTrack()
	writeJsonResponse(audios, err, w)
}

func nextSubtitles(w http.ResponseWriter, _ *http.Request) {
	subtitles, err := playerService.NextSubtitles()
	writeJsonResponse(subtitles, err, w)
}

func playMovie(w http.ResponseWriter, r *http.Request) {
	var entity struct{ Movie int `json:"movie"` }
	var status api.PlayerStatus
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		status, err = playerService.PlayMovie(entity.Movie)
	}
	writeJsonResponse(status, err, w)
}

func pause(w http.ResponseWriter, _ *http.Request) {
	status, err := playerService.Pause()
	writeJsonResponse(status, err, w)
}

func play(w http.ResponseWriter, _ *http.Request) {
	status, err := playerService.Play()
	writeJsonResponse(status, err, w)
}

func playPause(w http.ResponseWriter, _ *http.Request) {
	status, err := playerService.PlayPause()
	writeJsonResponse(status, err, w)
}

func previousAudioTrack(w http.ResponseWriter, _ *http.Request) {
	audios, err := playerService.PreviousAudioTrack()
	writeJsonResponse(audios, err, w)
}

func previousSubtitles(w http.ResponseWriter, _ *http.Request) {
	subtitles, err := playerService.PreviousSubtitles()
	writeJsonResponse(subtitles, err, w)
}

func replayCurrent(w http.ResponseWriter, _ *http.Request) {
	status, err := playerService.ReplayCurrent()
	writeJsonResponse(status, err, w)
}

func refresh(w http.ResponseWriter, _ *http.Request) {
	var err error
	conf, err = config.LoadConfig()
	if err == nil {
		catalogService.Refresh(conf)
	}
	writeJsonResponse(nil, err, w)
}

func searchMovies(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("q")
	writeJsonResponse(catalogService.Find(title), nil, w)
}

type index struct {
	Index int `json:"index"`
}

type position struct {
	Position int `json:"position"`
}

func seek(w http.ResponseWriter, r *http.Request) {
	var entity position
	parser := json.NewDecoder(r.Body)
	var status api.PlayerStatus
	var err error
	if err = parser.Decode(&entity); err == nil {
		status, err = playerService.Seek(entity.Position)
	}
	writeJsonResponse(status, err, w)
}

func selectAudio(w http.ResponseWriter, r *http.Request) {
	var entity index
	parser := json.NewDecoder(r.Body)
	var audios []omxcontrol.Stream
	var err error
	if err = parser.Decode(&entity); err == nil {
		audios, err = playerService.SelectAudio(entity.Index)
	}
	writeJsonResponse(audios, err, w)
}

func selectSubtitle(w http.ResponseWriter, r *http.Request) {
	var entity index
	parser := json.NewDecoder(r.Body)
	var subtitles []omxcontrol.Stream
	var err error
	if err = parser.Decode(&entity); err == nil {
		subtitles, err = playerService.SelectSubtitle(entity.Index)
	}
	writeJsonResponse(subtitles, err, w)
}

func setPosition(w http.ResponseWriter, r *http.Request) {
	var entity position
	parser := json.NewDecoder(r.Body)
	var status api.PlayerStatus
	var err error
	if err = parser.Decode(&entity); err == nil {
		status, err = playerService.SetPosition(entity.Position)
	}
	writeJsonResponse(status, err, w)
}

func status(w http.ResponseWriter, _ *http.Request) {
	st, err := playerService.Status()
	writeJsonResponse(st, err, w)
}

func stop(w http.ResponseWriter, _ *http.Request) {
	status, err := playerService.Stop()
	writeJsonResponse(status, err, w)
}

func subtitles(w http.ResponseWriter, _ *http.Request) {
	subtitles, err := playerService.Subtitles()
	writeJsonResponse(subtitles, err, w)
}

func toggleSubtitles(w http.ResponseWriter, _ *http.Request) {
	err := playerService.ToggleSubtitles()
	writeJsonResponse(nil, err, w)
}

func unmute(w http.ResponseWriter, _ *http.Request) {
	err := playerService.Unmute()
	writeJsonResponse(nil, err, w)
}

type vol struct {
	Volume float64 `json:"volume"`
}

func volume(w http.ResponseWriter, _ *http.Request) {
	v, err := playerService.Volume()
	writeJsonResponse(vol{v}, err, w)
}

func volumeDown(w http.ResponseWriter, _ *http.Request) {
	v, err := playerService.VolumeDown()
	writeJsonResponse(vol{v}, err, w)
}

func volumeUp(w http.ResponseWriter, _ *http.Request) {
	v, err := playerService.VolumeUp()
	writeJsonResponse(vol{v}, err, w)
}

func writeJsonResponse(body interface{}, err error, w http.ResponseWriter) {
	if body == nil && err == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err == nil {
		encoder.Encode(body)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		m := api.MessagePayload{Message: err.Error()}
		encoder.Encode(m)
	}
}
