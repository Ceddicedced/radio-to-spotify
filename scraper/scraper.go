package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type Song struct {
	Artist string
	Title  string
}

type Station struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Type      string        `json:"type"`
	ArtistTag string        `json:"artistTag,omitempty"`
	TitleTag  string        `json:"titleTag,omitempty"`
	ArtistKey []interface{} `json:"artistKey,omitempty"`
	TitleKey  []interface{} `json:"titleKey,omitempty"`
	Regex     string        `json:"regex,omitempty"`
}

type Config struct {
	Stations []Station `json:"stations"`
}

type StationSong struct {
	StationID string
	Station   string
	Song      Song
}

type Scraper interface {
	GetNowPlaying() (*Song, error)
}

type BaseScraper struct {
	Logger *logrus.Logger
	URL    string
}

func NewBaseScraper(logger *logrus.Logger, URL string) *BaseScraper {
	return &BaseScraper{Logger: logger, URL: URL}
}

func loadConfig(stationFile string) (*Config, error) {
	file, err := os.Open(stationFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func FetchNowPlaying(stationFile string, logger *logrus.Logger, stationID string) ([]*StationSong, error) {
	config, err := loadConfig(stationFile)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if stationID != "" {
		found := false
		for _, station := range config.Stations {
			if station.ID == stationID {
				stations = append(stations, station)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("no station found with id: %s", stationID)
		}
	} else {
		stations = config.Stations
	}

	var wg sync.WaitGroup
	results := make(chan *StationSong, len(stations))

	for _, station := range stations {
		wg.Add(1)
		go fetchStation(station, logger, &wg, results)
	}

	wg.Wait()
	close(results)

	var stationSongs []*StationSong
	for result := range results {
		stationSongs = append(stationSongs, result)
	}

	return stationSongs, nil
}

func fetchStation(station Station, logger *logrus.Logger, wg *sync.WaitGroup, results chan<- *StationSong) {
	defer wg.Done()

	var scraperInstance Scraper
	var err error

	switch station.Type {
	case "html":
		scraperInstance = NewHTMLScraper(logger, station.URL, station.ArtistTag, station.TitleTag)
	case "json":
		scraperInstance = NewJSONScraper(logger, station.URL, station.ArtistKey, station.TitleKey)
	case "plaintext":
		scraperInstance, err = NewPlaintextScraper(logger, station.URL, station.Regex)
		if err != nil {
			logger.Errorf("Error creating plaintext scraper for station %s (%s): %v", station.Name, station.ID, err)
			return
		}
	default:
		logger.Errorf("Unknown scraper type for station %s (%s): %s", station.Name, station.ID, station.Type)
		return
	}

	logger.Debugf("Fetching now playing for station: %s (%s)", station.Name, station.ID)
	nowPlaying, err := scraperInstance.GetNowPlaying()
	if err != nil {
		logger.Warnf("Error fetching now playing for station %s (%s): %v", station.Name, station.ID, err)
		return
	}

	results <- &StationSong{
		StationID: station.ID,
		Station:   station.Name,
		Song:      *nowPlaying,
	}
}
