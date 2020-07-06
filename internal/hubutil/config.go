package hubutil

import (
	"encoding/json"
	"io/ioutil"
)

type HubConfig struct {
	NotifyEmail  string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
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
	return cfg, nil
}
