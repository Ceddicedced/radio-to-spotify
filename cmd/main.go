package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"radio-to-spotify/scraper"
	"radio-to-spotify/spotify"
	"radio-to-spotify/storage"
)

var (
	logger = logrus.New()
	store  storage.Storage
)

func main() {
	// Define the subcommands
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
	storeCmd := flag.NewFlagSet("store", flag.ExitOnError)
	playlistCmd := flag.NewFlagSet("playlist", flag.ExitOnError)

	// Define common flags
	storageType := flag.String("storage", "file", "Storage type: file or sqlite")
	storagePath := flag.String("storage-path", "data.db", "Path to storage file or database")

	// Define flags for fetch command
	fetchConfigFile := fetchCmd.String("config", "config.json", "Path to config file")
	fetchStationID := fetchCmd.String("station", "", "Station ID to fetch now playing")
	fetchLogLevel := fetchCmd.String("loglevel", "info", "Logging level: debug, info, warn, error, fatal, panic")

	// Define flags for store command
	storeConfigFile := storeCmd.String("config", "config.json", "Path to config file")
	storeStationID := storeCmd.String("station", "", "Station ID to store now playing")
	storeLogLevel := storeCmd.String("loglevel", "info", "Logging level: debug, info, warn, error, fatal, panic")
	storeDryRun := storeCmd.Bool("dry-run", false, "Dry run mode")

	// Define flags for playlist command
	playlistConfigFile := playlistCmd.String("config", "config.json", "Path to config file")
	playlistStationID := playlistCmd.String("station", "", "Station ID to create playlist")
	playlistLogLevel := playlistCmd.String("loglevel", "info", "Logging level: debug, info, warn, error, fatal, panic")

	if len(os.Args) < 2 {
		fmt.Println("expected 'fetch', 'store' or 'playlist' subcommands")
		os.Exit(1)
	}

	// Parse common flags
	flag.Parse()

	var err error
	store, err = storage.NewStorage(*storageType, *storagePath)
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	switch os.Args[1] {
	case "fetch":
		fetchCmd.Parse(os.Args[2:])
		executeFetch(*fetchConfigFile, *fetchStationID, *fetchLogLevel)
	case "store":
		storeCmd.Parse(os.Args[2:])
		executeStore(*storeConfigFile, *storeStationID, *storeLogLevel, *storeDryRun)
	case "playlist":
		playlistCmd.Parse(os.Args[2:])
		executePlaylist(*playlistConfigFile, *playlistStationID, *playlistLogLevel)
	default:
		fmt.Println("expected 'fetch', 'store' or 'playlist' subcommands")
		os.Exit(1)
	}
}

func setLogLevel(logLevel string) {
	level, err := logrus.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		logger.Fatalf("Invalid log level: %v", err)
	}
	logger.SetLevel(level)
}

func executeFetch(configFile, stationID, logLevel string) {
	setLogLevel(logLevel)

	stationSongs, err := scraper.FetchNowPlaying(configFile, logger, stationID)
	if err != nil {
		logger.Fatalf("Error fetching now playing: %v", err)
	}

	for _, stationSong := range stationSongs {
		fmt.Printf("Station: %s (%s) - Artist: %s, Title: %s\n", stationSong.Station, stationSong.StationID, stationSong.Song.Artist, stationSong.Song.Title)
	}
}

func executeStore(configFile, stationID, logLevel string, dryRun bool) {
	setLogLevel(logLevel)

	stationSongs, err := scraper.FetchNowPlaying(configFile, logger, stationID)
	if err != nil {
		logger.Fatalf("Error fetching now playing: %v", err)
	}

	for _, stationSong := range stationSongs {
		if dryRun {
			fmt.Printf("Dry run: would store song for station %s: %s - %s\n", stationSong.StationID, stationSong.Song.Artist, stationSong.Song.Title)
		} else {
			err := store.StoreNowPlaying(stationSong.StationID, &stationSong.Song)
			if err != nil {
				logger.Fatalf("Error storing now playing: %v", err)
			}
			fmt.Printf("Stored song for station %s: %s - %s\n", stationSong.StationID, stationSong.Song.Artist, stationSong.Song.Title)
		}
	}
}

func executePlaylist(configFile, stationID, logLevel string) {
	setLogLevel(logLevel)

	err := spotify.CreateSpotifyPlaylist(stationID, store)
	if err != nil {
		logger.Fatalf("Error creating Spotify playlist: %v", err)
	}
}
