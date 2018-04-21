package api

type Movie struct {
	Available bool   `json:"available"`
	DriveName string `json:"drive"`
	Id        int    `json:"id"`
	Path      string `json:"path"`
	Title     string `json:"title"`
	TMDbId    int    `json:"tmdb_id,omitempty"`
}

type MovieDetails struct {
	Budget         int64    `json:"budget,omitempty"`
	Companies      []string `json:"companies,omitempty"`
	Countries      []string `json:"countries,omitempty"`
	Genres         []string `json:"genres,omitempty"`
	OriginalTitle  string   `json:"original_title"`
	Overview       string   `json:"overview"`
	PosterSmallUrl string   `json:"poster_small_url"`
	PosterLargeUrl string   `json:"poster_large_url"`
	ReleaseDate    string   `json:"release_date"`
	Revenue        int64    `json:"revenue,omitempty"`
	TagLine        string   `json:"tagline,omitempty"`
	TMDbId         int      `json:"tmdb_id"`
	ImdbId         string   `json:"imdb_id,omitempty"`
}

type PlayerStatus struct {
	Playing  string `json:"playing"`
	Duration int    `json:"duration"`
	Position int    `json:"position"`
	Paused   bool   `json:"paused"`
}

type MessagePayload struct {
	Message string `json:"message,omitempty"`
}
