package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"net/http"
	"os/signal"
	"syscall"
	"strconv"
	"time"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/gomovies/service"
	"github.com/andrew00x/omxcontrol"
)

var conf *config.Config
var catalogService *service.CatalogService
var playerService *service.PlayerService
var tmDbService *service.TMDbService

func main() {
	var err error
	conf, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not read configuration: %v", err)
	}

	catalogService, err = service.CreateCatalogService(conf)
	if err != nil {
		log.Fatalf("Could not create catalog: %v", err)
	}

	playerService, err = service.CreatePlayerService(conf)
	if err != nil {
		log.Fatalf("Could not create player: %v", err)
	}

	tmDbService, err = service.CreateTMDbService(conf)
	if err != nil {
		log.Fatalf("Could not the Movie DB service: %v", err)
	}

	log.Println("Start loading details from 'The Movie DB'")
	startTMDbLoad := time.Now()
	for _, m := range catalogService.All() {
		if m.TMDbId > 0 {
			tmDbService.MovieDetails(m.TMDbId)
		}
	}
	stopTMDbLoad := time.Now()
	log.Printf("Stop loading details from 'The Movie DB', spent %ds\n", stopTMDbLoad.Sub(startTMDbLoad)/time.Second)

	web := http.Server{Addr: fmt.Sprintf(":%d", conf.WebPort), Handler: http.DefaultServeMux}
	log.Printf("Starting on port: %d\n", conf.WebPort)

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		if err := catalogService.Save(); err != nil {
			log.Printf("Unable save catalog file: %v\n", err)
		} else {
			log.Println("Save catalog file")
		}
		if err := web.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not shutdown: %v", err)
		}
	}()

	http.HandleFunc("/api/details", details)
	http.HandleFunc("/api/details/search", searchDetails)
	http.HandleFunc("/api/list", allMovies)
	http.HandleFunc("/api/play", playMovie)
	http.HandleFunc("/api/search", searchMovies)
	http.HandleFunc("/api/refresh", refresh)
	http.HandleFunc("/api/update", updateMovie)
	http.HandleFunc("/api/player/audios", audios)
	http.HandleFunc("/api/player/enqueue", enqueue)
	http.HandleFunc("/api/player/dequeue", dequeue)
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
	webClient()

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
	result := catalogService.All()
	l := len(result)
	for i := 0; i < l; i++ {
		m := &result[i]
		if m.TMDbId > 0 {
			md, err := tmDbService.MovieDetails(m.TMDbId)
			if err == nil {
				m.Details = &md
			}
		}
	}
	writeJsonResponse(result, nil, w)
}

func audios(w http.ResponseWriter, _ *http.Request) {
	audios, err := playerService.AudioTracks()
	writeJsonResponse(audios, err, w)
}

func details(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	var md api.MovieDetails
	if err == nil {
		m, found := catalogService.Get(int(id))
		if found {
			if m.TMDbId != 0 {
				md, err = tmDbService.MovieDetails(m.TMDbId)
			}
		} else {
			err = errors.New(fmt.Sprintf("invalid movie id: %d", id))
		}
	}
	writeJsonResponse(md, err, w)
}

func enqueue(w http.ResponseWriter, r *http.Request) {
	var entity movie
	var queue []string
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		queue, err = playerService.Enqueue(entity.MoviePath)
	}
	writeJsonResponse(queue, err, w)
}

func dequeue(w http.ResponseWriter, r *http.Request) {
	var entity position
	var queue []string
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		queue = playerService.Dequeue(entity.Position)
	}
	writeJsonResponse(queue, err, w)
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
	var entity movie
	var status api.PlayerStatus
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		status, err = playerService.PlayMovie(entity.MoviePath)
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
	err := catalogService.Refresh()
	writeJsonResponse(nil, err, w)
}

func searchDetails(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	result, err := tmDbService.SearchDetails(query)
	writeJsonResponse(result, err, w)
}

func searchMovies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	result := catalogService.Find(query)
	for _, m := range result {
		if m.TMDbId > 0 {
			md, err := tmDbService.MovieDetails(m.TMDbId)
			if err == nil {
				m.Details = &md
			}
		}
	}
	writeJsonResponse(result, nil, w)
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
	var entity trackIndex
	parser := json.NewDecoder(r.Body)
	var audios []omxcontrol.Stream
	var err error
	if err = parser.Decode(&entity); err == nil {
		audios, err = playerService.SelectAudio(entity.Index)
	}
	writeJsonResponse(audios, err, w)
}

func selectSubtitle(w http.ResponseWriter, r *http.Request) {
	var entity trackIndex
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

func updateMovie(w http.ResponseWriter, r *http.Request) {
	var movie api.Movie
	parser := json.NewDecoder(r.Body)
	err := parser.Decode(&movie)
	if err == nil {
		movie, err = catalogService.Update(movie)
	}
	writeJsonResponse(movie, err, w)
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

type movie struct {
	MoviePath string `json:"movie"`
}

type trackIndex struct {
	Index int `json:"index"`
}

type position struct {
	Position int `json:"position"`
}

type vol struct {
	Volume float64 `json:"volume"`
}
