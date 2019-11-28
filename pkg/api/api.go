package api

type Movie struct {
	Available bool          `json:"available"`
	DriveName string        `json:"drive"`
	Id        int           `json:"id"`
	File      string        `json:"file"`
	Title     string        `json:"title"`
	TMDbId    int           `json:"tmdb_id,omitempty"`
	Details   *MovieDetails `json:"details,omitempty"`
}

type MovieDetails struct {
	Budget         int64    `json:"budget,omitempty"`
	Companies      []string `json:"companies,omitempty"`
	Countries      []string `json:"countries,omitempty"`
	Genres         []string `json:"genres,omitempty"`
	OriginalTitle  string   `json:"originalTitle"`
	Overview       string   `json:"overview,omitempty"`
	PosterSmallUrl string   `json:"posterSmallUrl"`
	PosterLargeUrl string   `json:"posterLargeUrl"`
	ReleaseDate    string   `json:"releaseDate"`
	Revenue        int64    `json:"revenue,omitempty"`
	TagLine        string   `json:"tagline,omitempty"`
	TMDbId         int      `json:"tmdbId"`
	ImdbId         string   `json:"imdbId,omitempty"`
}

type Playback struct {
	File             string `json:"file"`
	Position         int    `json:"position"`
	ActiveAudioTrack int    `json:"activeAudioTrack"`
	ActiveSubtitle   int    `json:"activeSubtitle"`
}

type PlayerStatus struct {
	File             string `json:"file"`
	Duration         int    `json:"duration"`
	Position         int    `json:"position"`
	Paused           bool   `json:"paused"`
	Muted            bool   `json:"muted"`
	SubtitlesOff     bool   `json:"subtitlesOff"`
	ActiveAudioTrack int    `json:"activeAudioTrack"`
	ActiveSubtitle   int    `json:"activeSubtitle"`
	Stopped          bool   `json:"stopped"`
}

type Stream struct {
	Index    int    `json:"index"`
	Language string `json:"lang"`
	Name     string `json:"name"`
	Codec    string `json:"codec"`
	Active   bool   `json:"active"`
}

type TorrentDownload struct {
	Name          string            `json:"name"`
	Path          string            `json:"path"`
	Size          int64             `json:"size"`
	CompletedSize int64             `json:"completedSize"`
	Completed     bool              `json:"completed"`
	Ratio         float32           `json:"ratio"`
	Attrs         map[string]string `json:"attrs,omitempty"`
}

type TorrentDownloadFile struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type MessagePayload struct {
	Message string `json:"message,omitempty"`
}

type MoviePath struct {
	File string `json:"file"`
}

type TrackIndex struct {
	Index int `json:"index"`
}

type Position struct {
	Position int `json:"position"`
}

type Volume struct {
	Volume float64 `json:"volume"`
}

type TorrentFile struct {
	Content string `json:"file"`
}
