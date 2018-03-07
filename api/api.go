package api

type Movie struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	DriveName string `json:"drive"`
	Available bool   `json:"available"`
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
