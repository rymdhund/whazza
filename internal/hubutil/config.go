package hubutil

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

type HubConfig struct {
	DataDir string `json:"data_dir"`
	// Mail settings
	NotifyEmail  string `json:"notify_email"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	SMTPFrom     string `json:"smtp_from"`
}

func (cfg HubConfig) Database() string {
	return path.Join(cfg.DataDir, "whazza.db")
}

func (cfg HubConfig) CertFile() string {
	return path.Join(cfg.DataDir, "cert.pem")
}

func (cfg HubConfig) KeyFile() string {
	return path.Join(cfg.DataDir, "key.pem")
}

func ReadConfig(filename string) (HubConfig, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return HubConfig{}, err
	}

	var cfg HubConfig
	err = json.Unmarshal(contents, &cfg)
	if err != nil {
		return HubConfig{}, err
	}

	if cfg.DataDir == "" {
		cfg.DataDir = path.Join(os.Getenv("HOME"), ".whazza")
	}
	return cfg, nil
}
