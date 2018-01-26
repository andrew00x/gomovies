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

	ctl, err := catalog.NewCatalog(conf)
	if err != nil {
		log.Fatalf("Could not create catalog: %v", err)
	}
	catalogService = service.NewCatalogService(ctl)

	var plr player.Player
	plr, err = player.Create(conf)
	if err != nil {
		log.Fatalf("Could not create player: %v", err)
	}
	playerService = service.CreatePlayerService(plr, ctl)

	web := http.Server{Addr: fmt.Sprintf(":%d", conf.WebPort), Handler: http.DefaultServeMux}
	log.Printf("Starting on port: %d\n", conf.WebPort)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		if err := catalogService.Stop(); err != nil {
			log.Printf("Unable save catalog file: %v\n", err)
		}
		if err := web.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not shutdown: %v", err)
		}
	}()

	webClient()
	http.HandleFunc("/api/list", allMovies)
	http.HandleFunc("/api/search", searchMovies)
	http.HandleFunc("/api/play", startPlay)
	http.HandleFunc("/api/enqueue", enqueue)
	http.HandleFunc("/api/player/playpause", playPause)
	http.HandleFunc("/api/player/stop", stopPlay)
	http.HandleFunc("/api/player/replay", replayCurrent)
	http.HandleFunc("/api/player/forward30s", forward30s)
	http.HandleFunc("/api/player/rewind30s", rewind30s)
	http.HandleFunc("/api/player/forward10m", forward10m)
	http.HandleFunc("/api/player/rewind10m", rewind10m)
	http.HandleFunc("/api/player/volumeup", volumeUp)
	http.HandleFunc("/api/player/volumedown", volumeDown)
	http.HandleFunc("/api/player/nextaudiotrack", nextAudioTrack)
	http.HandleFunc("/api/player/previousaudiotrack", previousAudioTrack)
	http.HandleFunc("/api/player/nextsubtitles", nextSubtitles)
	http.HandleFunc("/api/player/previoussubtitles", previousSubtitles)
	http.HandleFunc("/api/player/togglesubtitles", toggleSubtitles)
	http.HandleFunc("/api/player/playlist", playlist)
	http.HandleFunc("/api/player/playlist/next", nextInPlaylist)
	http.HandleFunc("/api/player/playlist/previous", previousInPlaylist)
	http.HandleFunc("/api/player/playlist/delete", deleteInPlaylist)
	http.HandleFunc("/api/player/playlist/play", playInPlaylist)
	http.HandleFunc("/api/player/status", status)

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

func searchMovies(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("q")
	writeJsonResponse(catalogService.Find(name), nil, w)
}

func status(w http.ResponseWriter, _ *http.Request) {
	st, err := playerService.Status()
	writeJsonResponse(st, err, w)
}

func startPlay(w http.ResponseWriter, r *http.Request) {
	var entity struct{ Movie int `json:"movie"` }
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		err = playerService.Play(entity.Movie)
	}
	writeJsonResponse(nil, err, w)
}

func enqueue(w http.ResponseWriter, r *http.Request) {
	var entity struct{ Movie int `json:"movie"` }
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		err = playerService.Enqueue(entity.Movie)
	}
	writeJsonResponse(nil, err, w)
}

func playPause(w http.ResponseWriter, _ *http.Request) {
	err := playerService.PlayPause()
	writeJsonResponse(nil, err, w)
}

func stopPlay(w http.ResponseWriter, _ *http.Request) {
	err := playerService.StopPlay()
	writeJsonResponse(nil, err, w)
}

func replayCurrent(w http.ResponseWriter, _ *http.Request) {
	err := playerService.ReplayCurrent()
	writeJsonResponse(nil, err, w)
}

func forward30s(w http.ResponseWriter, _ *http.Request) {
	err := playerService.Forward30s()
	writeJsonResponse(nil, err, w)
}

func rewind30s(w http.ResponseWriter, _ *http.Request) {
	err := playerService.Rewind30s()
	writeJsonResponse(nil, err, w)
}

func forward10m(w http.ResponseWriter, _ *http.Request) {
	err := playerService.Forward10m()
	writeJsonResponse(nil, err, w)
}

func rewind10m(w http.ResponseWriter, _ *http.Request) {
	err := playerService.Rewind10m()
	writeJsonResponse(nil, err, w)
}

func volumeUp(w http.ResponseWriter, _ *http.Request) {
	err := playerService.VolumeUp()
	writeJsonResponse(nil, err, w)
}

func volumeDown(w http.ResponseWriter, _ *http.Request) {
	err := playerService.VolumeDown()
	writeJsonResponse(nil, err, w)
}

func nextAudioTrack(w http.ResponseWriter, _ *http.Request) {
	err := playerService.NextAudioTrack()
	writeJsonResponse(nil, err, w)
}

func previousAudioTrack(w http.ResponseWriter, _ *http.Request) {
	err := playerService.PreviousAudioTrack()
	writeJsonResponse(nil, err, w)
}

func nextSubtitles(w http.ResponseWriter, _ *http.Request) {
	err := playerService.NextSubtitles()
	writeJsonResponse(nil, err, w)
}

func previousSubtitles(w http.ResponseWriter, _ *http.Request) {
	err := playerService.PreviousSubtitles()
	writeJsonResponse(nil, err, w)
}

func toggleSubtitles(w http.ResponseWriter, _ *http.Request) {
	err := playerService.ToggleSubtitles()
	writeJsonResponse(nil, err, w)
}

func playlist(w http.ResponseWriter, _ *http.Request) {
	pl, err := playerService.Playlist()
	writeJsonResponse(pl, err, w)
}

func nextInPlaylist(w http.ResponseWriter, _ *http.Request) {
	err := playerService.NextInPlaylist()
	writeJsonResponse(nil, err, w)
}

func previousInPlaylist(w http.ResponseWriter, _ *http.Request) {
	err := playerService.PreviousInPlaylist()
	writeJsonResponse(nil, err, w)
}

func deleteInPlaylist(w http.ResponseWriter, r *http.Request) {
	var entity struct{ Pos int `json:"pos"` }
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		err = playerService.DeleteInPlaylist(entity.Pos)
	}
	writeJsonResponse(nil, err, w)
}

func playInPlaylist(w http.ResponseWriter, r *http.Request) {
	var entity struct{ Pos int `json:"pos"` }
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		err = playerService.PlayInPlaylist(entity.Pos)
	}
	writeJsonResponse(nil, err, w)
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
