package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	CatalogFile      string   `json:"catalog_file"`
	Dirs             []string `json:"dirs"`
	SavePlaybackTime string   `json:"save_playback_time"`
	VideoFileExts    []string `json:"video_file_exts"`
	WebDir           string   `json:"web_dir"`
	WebPort          int      `json:"web_port"`
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
	if conf.SavePlaybackTime == "" {
		conf.SavePlaybackTime = "yes"
	}
	if conf.WebPort == 0 {
		conf.WebPort = 8000
	}
	if conf.CatalogFile == "" {
		conf.CatalogFile = filepath.Join(filepath.Dir(path), "catalog.json")
	}
	return &conf, nil
}

func ConfDir() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".gomovies")
}
