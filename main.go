package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"radio-to-spotify-playlist/scraper"
	"sync"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Stations []scraper.Station `json:"stations"`
}

func loadConfig(configFile string) (*Config, error) {
	file, err := os.Open(configFile)
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

func fetchNowPlaying(station scraper.Station, logger *logrus.Logger, wg *sync.WaitGroup, results chan<- *scraper.Song) {
	defer wg.Done()

	var scraperInstance scraper.Scraper
	var err error

	switch station.Type {
	case "html":
		scraperInstance = scraper.NewHTMLScraper(logger, station.URL, station.ArtistTag, station.TitleTag)
	case "json":
		scraperInstance = scraper.NewJSONScraper(logger, station.URL, station.ArtistKey, station.TitleKey)
	case "plaintext":
		scraperInstance, err = scraper.NewPlaintextScraper(logger, station.URL, station.Regex)
		if err != nil {
			logger.Errorf("Error creating plaintext scraper for station %s: %v", station.Name, err)
			return
		}
	default:
		logger.Errorf("Unknown scraper type for station %s: %s", station.Name, station.Type)
		return
	}

	logger.Infof("Fetching now playing for station: %s", station.Name)
	nowPlaying, err := scraperInstance.GetNowPlaying()
	if err != nil {
		logger.Errorf("Error fetching now playing for station %s: %v", station.Name, err)
		return
	}

	logger.Infof("Now playing for station %s: %s - %s", station.Name, nowPlaying.Artist, nowPlaying.Title)

	results <- &scraper.Song{
		Time:   "",
		Artist: nowPlaying.Artist,
		Title:  nowPlaying.Title,
	}
}

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting the radio-to-spotify application")

	config, err := loadConfig("config.json")
	if err != nil {
		logger.Fatalf("Error loading config file: %v", err)
	}

	var wg sync.WaitGroup
	results := make(chan *scraper.Song, len(config.Stations))

	for _, station := range config.Stations {
		wg.Add(1)
		go fetchNowPlaying(station, logger, &wg, results)
	}

	wg.Wait()
	close(results)

	for result := range results {
		fmt.Printf("Artist: %s, Title: %s\n", result.Artist, result.Title)
	}

	logger.Info("Finished the radio-to-spotify application")
}
