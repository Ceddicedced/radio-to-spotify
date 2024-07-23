package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"radio-to-spotify/scraper"
	"radio-to-spotify/spotify"
	"radio-to-spotify/storage"
)

func main() {
	configFile := flag.String("config", "config.json", "Path to config file")
	stationID := flag.String("station", "", "Station ID to fetch now playing")
	action := flag.String("action", "fetch", "Action to perform: fetch, store, playlist")
	flag.Parse()

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	store := storage.NewStorage()

	switch *action {
	case "fetch":
		songs, err := scraper.FetchNowPlaying(*configFile, logger, *stationID)
		if err != nil {
			logger.Fatalf("Error fetching now playing: %v", err)
		}
		for _, song := range songs {
			fmt.Printf("Artist: %s, Title: %s\n", song.Artist, song.Title)
		}
	case "store":
		songs, err := scraper.FetchNowPlaying(*configFile, logger, *stationID)
		if err != nil {
			logger.Fatalf("Error fetching now playing: %v", err)
		}
		for _, song := range songs {
			err := store.StoreNowPlaying(*stationID, song)
			if err != nil {
				logger.Fatalf("Error storing now playing: %v", err)
			}
			fmt.Printf("Stored song for station %s: %s - %s\n", *stationID, song.Artist, song.Title)
		}
	case "playlist":
		err := spotify.CreateSpotifyPlaylist(*stationID, store)
		if err != nil {
			logger.Fatalf("Error creating Spotify playlist: %v", err)
		}
	default:
		logger.Fatalf("Unknown action: %s", *action)
	}
}
