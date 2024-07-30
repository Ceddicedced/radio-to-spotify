package scraper

import (
	"sync"

	"radio-to-spotify/config"

	"github.com/sirupsen/logrus"
)

type Song struct {
	Artist string
	Title  string
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

func FetchNowPlaying(configHandler *config.ConfigHandler, logger *logrus.Logger, stationID string) ([]*config.Station, []*Song, error) {
	var stations []config.Station
	if stationID != "" {
		station, err := configHandler.GetStationByID(stationID)
		if err != nil {
			return nil, nil, err
		}
		stations = append(stations, *station)
	} else {
		stations = configHandler.GetAllStations()
	}

	var wg sync.WaitGroup
	results := make(chan struct {
		Station *config.Station
		Song    *Song
	}, len(stations))

	for _, station := range stations {
		wg.Add(1)
		go fetchStation(&station, logger, &wg, results)
	}

	wg.Wait()
	close(results)

	var stationSongs []*config.Station
	var songs []*Song
	for result := range results {
		stationSongs = append(stationSongs, result.Station)
		songs = append(songs, result.Song)
	}

	return stationSongs, songs, nil
}

func fetchStation(station *config.Station, logger *logrus.Logger, wg *sync.WaitGroup, results chan<- struct {
	Station *config.Station
	Song    *Song
}) {
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

	logger.Infof("Fetching now playing for station: %s (%s)", station.Name, station.ID)
	nowPlaying, err := scraperInstance.GetNowPlaying()
	if err != nil {
		logger.Errorf("Error fetching now playing for station %s (%s): %v", station.Name, station.ID, err)
		return
	}

	results <- struct {
		Station *config.Station
		Song    *Song
	}{
		Station: station,
		Song:    nowPlaying,
	}
}
