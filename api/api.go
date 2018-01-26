package api

type Movie struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	DriveName string `json:"drive"`
	Available bool   `json:"available"`
}

type Playlist struct {
	Items []string `json:"items"`
}

type PlayerStatus struct {
	Playing  string    `json:"playing"`
	Duration int       `json:"duration"`
	Position int       `json:"position"`
	Paused   bool      `json:"paused"`
	Streams  []*Stream `json:"streams"`
}

type Stream struct {
	Num     int    `json:"num"`
	Lang    string `json:"lang"`
	Type    string `json:"type"`
	Default bool   `json:"def"`
}

type MessagePayload struct {
	Message string `json:"message,omitempty"`
}
