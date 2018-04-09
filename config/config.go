package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Dirs          []string `json:"dirs"`
	VideoFileExts []string `json:"video_file_exts"`
	WebDir        string   `json:"web_dir"`
	WebPort       int      `json:"web_port"`
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
	return &conf, nil
}

func ConfDir() string {
	confDir := os.Getenv("GO_MOVIES_HOME")
	if confDir == "" {
		confDir = filepath.Join(os.Getenv("HOME"), ".gomovies")
	}
	return confDir
}
