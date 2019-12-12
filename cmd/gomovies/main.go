package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/gomovies/pkg/service"
)

var conf *config.Config
var catalogService *service.CatalogService
var playerService *service.PlayerService
var detailsService *service.DetailsService
var torrentService *service.TorrentService
var detailsLoadedFlag int32

func isDetailsLoaded() (loaded bool) {
	if atomic.LoadInt32(&detailsLoadedFlag) != 0 {
		loaded = true
	}
	return
}

func setDetailsLoaded(value bool) {
	var f int32
	if value {
		f = 1
	}
	atomic.StoreInt32(&detailsLoadedFlag, f)
}

type errResponse struct {
	err  error
	code int
}

func newErrResponse(err error, code int) *errResponse {
	return &errResponse{err: err, code: code}
}

func isErrResponse(err error) bool {
	_, ok := err.(*errResponse)
	return ok
}

func (e *errResponse) Error() string {
	return e.err.Error()
}

func init() {
	v := flag.Bool("v", false, "makes logger be more verbose, debug level")
	flag.Parse()
	if *v {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	var err error
	conf, err = config.LoadConfig()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Could not read configuration")
	}

	catalogService, err = service.CreateCatalogService(conf)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Could not create catalog")
	}

	playerService, err = service.CreatePlayerService(conf)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Could not create player")
	}

	if conf.TorrentRemoteCtrlAddr != "" {
		torrentService = service.CreateTorrentService(conf)
	}

	detailsService, err = service.CreateDetailsService(conf)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Could not create movies' details service")
	}
	go loadDetails()

	server := http.Server{Addr: fmt.Sprintf(":%d", conf.WebPort), Handler: http.DefaultServeMux}
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		var err error
		if err = catalogService.Save(); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable save catalog file")
		} else {
			log.Info("Catalog file saved")
		}
		if err = server.Shutdown(context.Background()); err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Could not shutdown")
		}
	}()

	http.HandleFunc("/api/details", details)
	http.HandleFunc("/api/details/search", searchDetails)
	http.HandleFunc("/api/list", allMovies)
	http.HandleFunc("/api/play", playMovie)
	http.HandleFunc("/api/enqueue", enqueue)
	http.HandleFunc("/api/dequeue", dequeue)
	http.HandleFunc("/api/queue", queue)
	http.HandleFunc("/api/clearqueue", clearQueue)
	http.HandleFunc("/api/shiftqueue", shiftQueue)
	http.HandleFunc("/api/search", searchMovies)
	http.HandleFunc("/api/refresh", refresh)
	http.HandleFunc("/api/update", updateMovie)
	http.HandleFunc("/api/player/audios", audios)
	http.HandleFunc("/api/player/nextaudiotrack", nextAudioTrack)
	http.HandleFunc("/api/player/nextsubtitle", nextSubtitle)
	http.HandleFunc("/api/player/pause", pause)
	http.HandleFunc("/api/player/play", play)
	http.HandleFunc("/api/player/playpause", playPause)
	http.HandleFunc("/api/player/previousaudiotrack", previousAudioTrack)
	http.HandleFunc("/api/player/previoussubtitle", previousSubtitle)
	http.HandleFunc("/api/player/replay", replayCurrent)
	http.HandleFunc("/api/player/seek", seek)
	http.HandleFunc("/api/player/audio", selectAudio)
	http.HandleFunc("/api/player/subtitle", selectSubtitle)
	http.HandleFunc("/api/player/position", setPosition)
	http.HandleFunc("/api/player/status", status)
	http.HandleFunc("/api/player/stop", stop)
	http.HandleFunc("/api/player/subtitles", subtitles)
	http.HandleFunc("/api/player/togglemute", toggleMute)
	http.HandleFunc("/api/player/togglesubtitles", toggleSubtitles)
	http.HandleFunc("/api/player/volume", volume)
	http.HandleFunc("/api/player/volumedown", volumeDown)
	http.HandleFunc("/api/player/volumeup", volumeUp)
	http.HandleFunc("/api/torrent/add", torrentAddFile)
	http.HandleFunc("/api/torrent/list", torrentListDownloads)
	http.HandleFunc("/api/torrent/files", torrentListFiles)
	http.HandleFunc("/api/torrent/stop", torrentStop)
	http.HandleFunc("/api/torrent/start", torrentStart)
	http.HandleFunc("/api/torrent/delete", torrentDelete)

	log.WithFields(log.Fields{"port": conf.WebPort}).Info("Starting")
	if err = server.ListenAndServe(); err != http.ErrServerClosed {
		log.WithFields(log.Fields{"err": err}).Fatal("Could not start http listener")
	}
}

func loadDetails() {
	if isDetailsLoaded() {
		log.Info("Skip loading movies' details since they are already loaded")
		return
	}
	log.Info("Start loading movies' details")
	startDetailsLoad := time.Now()
	for _, m := range catalogService.All() {
		for _, lang := range conf.DetailsLangs {
			if d, ok, e := detailsService.MovieDetails(m, lang, true); e != nil {
				log.WithFields(log.Fields{"err": e, "movie": m.Title}).Warn("Error occurred while loading movie details")
			} else if ok {
				tags := []string{d.Title, d.OriginalTitle}
				for _, g := range d.Genres {
					tags = append(tags, g)
				}
				for _, t := range tags {
					e = catalogService.AddTag(t, m.Id)
					if e != nil {
						log.WithFields(log.Fields{"err": e, "movie": m.Title}).Warn("Error occurred while adding tag for movie")
					}
				}
			}
		}
	}
	stopDetailsLoad := time.Now()
	log.WithFields(log.Fields{
		"spent_time": stopDetailsLoad.Sub(startDetailsLoad).Truncate(time.Second),
	}).Info("Stop loading movies' details")
	setDetailsLoaded(true)
}

func allMovies(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	result := catalogService.All()
	l := len(result)
	for i := 0; i < l; i++ {
		m := &result[i]
		md, found, err := detailsService.MovieDetails(*m, lang, isDetailsLoaded())
		if err == nil && found {
			m.Details = &md
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
		lang := r.URL.Query().Get("lang")
		if lang == "" {
			lang = "en"
		}
		m, found := catalogService.Get(int(id))
		if found {
			md, _, err = detailsService.MovieDetails(m, lang, true)
		} else {
			err = newErrResponse(fmt.Errorf("invalid movie id: %d", id), 404)
		}
	}
	writeJsonResponse(md, err, w)
}

func enqueue(w http.ResponseWriter, r *http.Request) {
	var entity []api.MoviePath
	var queue []string
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		queue, err = playerService.Enqueue(unwrapFiles(entity))
	}
	writeJsonResponse(wrapFiles(queue), err, w)
}

func dequeue(w http.ResponseWriter, r *http.Request) {
	var entity api.Position
	var queue []string
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		queue = playerService.Dequeue(entity.Position)
	}
	writeJsonResponse(wrapFiles(queue), err, w)
}

func queue(w http.ResponseWriter, _ *http.Request) {
	writeJsonResponse(wrapFiles(playerService.Queue()), nil, w)
}

func clearQueue(_ http.ResponseWriter, _ *http.Request) {
	playerService.ClearQueue()
}

func shiftQueue(w http.ResponseWriter, r *http.Request) {
	var entity api.Position
	var queue []string
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		queue = playerService.ShiftQueue(entity.Position)
	}
	writeJsonResponse(wrapFiles(queue), err, w)
}

func wrapFiles(files []string) (paths []api.MoviePath) {
	paths = make([]api.MoviePath, len(files))
	for i, f := range files {
		paths[i] = api.MoviePath{File: f}
	}
	return
}

func unwrapFiles(paths []api.MoviePath) (files []string) {
	files = make([]string, len(paths))
	for i, p := range paths {
		files[i] = p.File
	}
	return
}

func nextAudioTrack(w http.ResponseWriter, _ *http.Request) {
	audios, err := playerService.NextAudioTrack()
	writeJsonResponse(audios, err, w)
}

func nextSubtitle(w http.ResponseWriter, _ *http.Request) {
	subtitles, err := playerService.NextSubtitle()
	writeJsonResponse(subtitles, err, w)
}

func playMovie(w http.ResponseWriter, r *http.Request) {
	var entity api.Playback
	var status api.PlayerStatus
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&entity); err == nil {
		status, err = playerService.PlayMovie(entity)
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

func previousSubtitle(w http.ResponseWriter, _ *http.Request) {
	subtitles, err := playerService.PreviousSubtitle()
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
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	result, err := detailsService.SearchDetails(query, lang)
	writeJsonResponse(result, err, w)
}

func searchMovies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	result := catalogService.Find(query)
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	l := len(result)
	for i := 0; i < l; i++ {
		m := &result[i]
		md, found, err := detailsService.MovieDetails(*m, lang, isDetailsLoaded())
		if err == nil && found {
			m.Details = &md
		}
	}
	writeJsonResponse(result, nil, w)
}

func seek(w http.ResponseWriter, r *http.Request) {
	var entity api.Position
	parser := json.NewDecoder(r.Body)
	var status api.PlayerStatus
	var err error
	if err = parser.Decode(&entity); err == nil {
		status, err = playerService.Seek(entity.Position)
	}
	writeJsonResponse(status, err, w)
}

func selectAudio(w http.ResponseWriter, r *http.Request) {
	var entity api.TrackIndex
	parser := json.NewDecoder(r.Body)
	var audios []api.Stream
	var err error
	if err = parser.Decode(&entity); err == nil {
		audios, err = playerService.SelectAudio(entity.Index)
	}
	writeJsonResponse(audios, err, w)
}

func selectSubtitle(w http.ResponseWriter, r *http.Request) {
	var entity api.TrackIndex
	parser := json.NewDecoder(r.Body)
	var subtitles []api.Stream
	var err error
	if err = parser.Decode(&entity); err == nil {
		subtitles, err = playerService.SelectSubtitle(entity.Index)
	}
	writeJsonResponse(subtitles, err, w)
}

func setPosition(w http.ResponseWriter, r *http.Request) {
	var entity api.Position
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
	st, err := playerService.ToggleSubtitles()
	writeJsonResponse(st, err, w)
}

func toggleMute(w http.ResponseWriter, _ *http.Request) {
	st, err := playerService.ToggleMute()
	writeJsonResponse(st, err, w)
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
	writeJsonResponse(api.Volume{Volume: v}, err, w)
}

func volumeDown(w http.ResponseWriter, _ *http.Request) {
	v, err := playerService.VolumeDown()
	writeJsonResponse(api.Volume{Volume: v}, err, w)
}

func volumeUp(w http.ResponseWriter, _ *http.Request) {
	v, err := playerService.VolumeUp()
	writeJsonResponse(api.Volume{Volume: v}, err, w)
}

func torrentAddFile(w http.ResponseWriter, r *http.Request) {
	var torrent api.TorrentFile
	var err error
	parser := json.NewDecoder(r.Body)
	if err = parser.Decode(&torrent); err == nil {
		var file []byte
		if file, err = base64.StdEncoding.DecodeString(torrent.Content); err == nil {
			err = torrentService.AddFile(file)
		}
	}
	writeJsonResponse(nil, err, w)
}

func torrentListDownloads(w http.ResponseWriter, _ *http.Request) {
	d, err := torrentService.Torrents()
	writeJsonResponse(d, err, w)
}

func torrentListFiles(w http.ResponseWriter, r *http.Request) {
	var d api.TorrentDownload
	var err error
	var files []api.TorrentDownloadFile
	if d, err = parseTorrentDownload(r); err != nil {
		files, err = torrentService.Files(d)
	}
	writeJsonResponse(files, err, w)
}

func torrentStop(w http.ResponseWriter, r *http.Request) {
	var d api.TorrentDownload
	var err error
	if d, err = parseTorrentDownload(r); err != nil {
		err = torrentService.Stop(d)
	}
	writeJsonResponse(nil, err, w)
}

func torrentStart(w http.ResponseWriter, r *http.Request) {
	var d api.TorrentDownload
	var err error
	if d, err = parseTorrentDownload(r); err != nil {
		err = torrentService.Start(d)
	}
	writeJsonResponse(nil, err, w)
}

func torrentDelete(w http.ResponseWriter, r *http.Request) {
	var d api.TorrentDownload
	var err error
	if d, err = parseTorrentDownload(r); err != nil {
		err = torrentService.Delete(d)
	}
	writeJsonResponse(nil, err, w)
}

func parseTorrentDownload(r *http.Request) (d api.TorrentDownload, err error) {
	parser := json.NewDecoder(r.Body)
	err = parser.Decode(&d)
	return
}

func writeJsonResponse(body interface{}, err error, w http.ResponseWriter) {
	if body == nil && err == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err == nil {
		if e := encoder.Encode(body); e != nil {
			log.WithFields(log.Fields{"err": err}).Error("Error occurred while write response")
		}
	} else {
		if isErrResponse(err) {
			w.WriteHeader(err.(*errResponse).code)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		m := api.MessagePayload{Message: err.Error()}
		if e := encoder.Encode(m); e != nil {
			log.WithFields(log.Fields{"err": err}).Error("Error occurred while write response")
		}
	}
}
