package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Dirs            []string `json:"dirs"`
	VideoFileExts   []string `json:"video_file_exts"`
	WebDir          string   `json:"web_dir"`
	WebPort         int      `json:"web_port"`
	TMDbApiKey      string   `json:"tmdb_api_key"`
	TMDbPosterSmall string   `json:"tmdb_poster_small"`
	TMDbPosterLarge string   `json:"tmdb_poster_large"`
}

func LoadConfig() (*Config, error) {
	return loadConfig(filepath.Join(ConfDir(), "config.json"))
}

func loadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	parser := json.NewDecoder(configFile)
	var conf Config
	if err = parser.Decode(&conf); err != nil {
		return nil, err
	}
	if conf.WebPort == 0 {
		conf.WebPort = 8000
	}
	if len(conf.VideoFileExts) == 0 {
		conf.VideoFileExts = append(conf.VideoFileExts, ".avi", ".mkv")
	}
	if conf.TMDbPosterSmall == "" {
		conf.TMDbPosterSmall = "w154"
	}
	if conf.TMDbPosterLarge == "" {
		conf.TMDbPosterLarge = "w500"
	}
	return &conf, nil
}

func ConfDir() string {
	confDir := os.Getenv("GO_MOVIES_HOME")
	if confDir == "" {
		confDir = filepath.Join(os.Getenv("HOME"), ".gomovies")
	}
	return confDir
}
