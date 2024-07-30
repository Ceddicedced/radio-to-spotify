package cmd

import (
	"fmt"
	"radio-to-spotify/config"
	"radio-to-spotify/scraper"
	"radio-to-spotify/storage"

	"github.com/spf13/cobra"
)

var storeDryRun bool

func init() {
	storeCmd.Flags().BoolVar(&storeDryRun, "dry-run", false, "Dry run mode")
	rootCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Store now playing songs from radio stations",
	Run: func(cmd *cobra.Command, args []string) {
		executeStore()
	},
}

func executeStore() {
	configHandler, err := config.NewConfigHandler(stationFile)
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	store, err := storage.NewStorage(storageType, storagePath)
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	stationSongs, songs, err := scraper.FetchNowPlaying(configHandler, logger, stationID)
	if err != nil {
		logger.Fatalf("Error fetching now playing: %v", err)
	}

	for i, station := range stationSongs {
		if storeDryRun {
			fmt.Printf("Dry run: would store song for station %s: %s - %s\n", station.ID, songs[i].Artist, songs[i].Title)
		} else {
			err := store.StoreNowPlaying(station.ID, songs[i])
			if err != nil {
				logger.Fatalf("Error storing now playing for station %s: %v", station.ID, err)
			}
			fmt.Printf("Stored song for station %s: %s - %s\n", station.ID, songs[i].Artist, songs[i].Title)
		}
	}
}
