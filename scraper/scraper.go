package scraper

import (
	"radio-to-spotify/utils"
	"sync"
)

type Song struct {
	Artist string
	Title  string
}

type Scraper interface {
	GetNowPlaying() (*Song, error)
}

type BaseScraper struct {
	URL string
}

func NewBaseScraper(URL string) *BaseScraper {
	return &BaseScraper{URL: URL}
}

func FetchNowPlaying(configHandler *utils.ConfigHandler, stationID string) ([]*utils.Station, []*Song, error) {
	var stations []utils.Station
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
		Station *utils.Station
		Song    *Song
	}, len(stations))

	for _, station := range stations {
		wg.Add(1)
		go fetchStation(&station, &wg, results)
	}

	wg.Wait()
	close(results)

	var stationSongs []*utils.Station
	var songs []*Song
	for result := range results {
		stationSongs = append(stationSongs, result.Station)
		songs = append(songs, result.Song)
	}

	return stationSongs, songs, nil
}

func fetchStation(station *utils.Station, wg *sync.WaitGroup, results chan<- struct {
	Station *utils.Station
	Song    *Song
}) {
	defer wg.Done()

	var scraperInstance Scraper
	var err error

	switch station.Type {
	case "html":
		scraperInstance = NewHTMLScraper(station.URL, station.ArtistTag, station.TitleTag)
	case "json":
		scraperInstance = NewJSONScraper(station.URL, station.ArtistKey, station.TitleKey)
	case "plaintext":
		scraperInstance, err = NewPlaintextScraper(station.URL, station.Regex)
		if err != nil {
			utils.Logger.Errorf("Error creating plaintext scraper for station %s (%s): %v", station.Name, station.ID, err)
			return
		}
	default:
		utils.Logger.Errorf("Unknown scraper type for station %s (%s): %s", station.Name, station.ID, station.Type)
		return
	}

	utils.Logger.Debugf("Fetching now playing for station: %s (%s)", station.Name, station.ID)
	nowPlaying, err := scraperInstance.GetNowPlaying()
	if err != nil {
		utils.Logger.Warnf("Error fetching now playing for station %s (%s): %v", station.Name, station.ID, err)
		return
	}

	results <- struct {
		Station *utils.Station
		Song    *Song
	}{
		Station: station,
		Song:    nowPlaying,
	}
}
