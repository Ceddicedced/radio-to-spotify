package config

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type Station struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	URL        string        `json:"url"`
	Type       string        `json:"type"`
	ArtistTag  string        `json:"artistTag,omitempty"`
	TitleTag   string        `json:"titleTag,omitempty"`
	ArtistKey  []interface{} `json:"artistKey,omitempty"`
	TitleKey   []interface{} `json:"titleKey,omitempty"`
	Regex      string        `json:"regex,omitempty"`
	PlaylistID string        `json:"playlistId,omitempty"`
}

type Config struct {
	Stations []Station `json:"stations"`
}

type ConfigHandler struct {
	mu       sync.Mutex
	filePath string
	config   *Config
}

func NewConfigHandler(filePath string) (*ConfigHandler, error) {
	handler := &ConfigHandler{
		filePath: filePath,
	}
	err := handler.load()
	if err != nil {
		return nil, err
	}
	return handler, nil
}

func (h *ConfigHandler) load() error {
	file, err := os.Open(h.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return err
	}

	h.config = &config
	return nil
}

func (h *ConfigHandler) save() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	file, err := os.Create(h.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(h.config)
}

func (h *ConfigHandler) GetStationByID(id string) (*Station, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, station := range h.config.Stations {
		if station.ID == id {
			return &h.config.Stations[i], nil
		}
	}
	return nil, errors.New("station not found")
}

func (h *ConfigHandler) GetAllStations() []Station {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.config.Stations
}

func (h *ConfigHandler) UpdateStation(station *Station) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, s := range h.config.Stations {
		if s.ID == station.ID {
			h.config.Stations[i] = *station
			return h.save()
		}
	}
	return errors.New("station not found")
}
