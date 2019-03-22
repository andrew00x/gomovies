package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Dirs                  []string `json:"dirs"`
	VideoFileExts         []string `json:"video_file_exts"`
	WebDir                string   `json:"web_dir"`
	WebPort               int      `json:"web_port"`
	TMDbApiKey            string   `json:"tmdb_api_key"`
	TMDbPosterSmall       string   `json:"tmdb_poster_small"`
	TMDbPosterLarge       string   `json:"tmdb_poster_large"`
	TorrentRemoteCtrlAddr string   `json:"torrent_remote_ctrl_addr"`
}

func LoadConfig() (*Config, error) {
	return loadConfig(filepath.Join(ConfDir(), "config.json"))
}

func loadConfig(path string) (conf *Config, err error) {
	configFile, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		if clsErr := configFile.Close(); clsErr != nil {
			err = clsErr
		}
	}()
	conf = &Config{}
	parser := json.NewDecoder(configFile)
	if err = parser.Decode(conf); err != nil {
		return
	}
	if conf.WebPort == 0 {
		conf.WebPort = 8000
	}
	if len(conf.VideoFileExts) == 0 {
		conf.VideoFileExts = append(conf.VideoFileExts, ".avi", ".mkv")
	}
	if conf.TMDbPosterSmall == "" {
		conf.TMDbPosterSmall = "w92"
	}
	if conf.TMDbPosterLarge == "" {
		conf.TMDbPosterLarge = "w500"
	}
	return
}

func ConfDir() string {
	confDir := os.Getenv("GO_MOVIES_HOME")
	if confDir == "" {
		confDir = filepath.Join(os.Getenv("HOME"), ".gomovies")
	}
	return confDir
}
