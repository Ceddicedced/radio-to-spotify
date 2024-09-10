package cmd

import (
	"radio-to-spotify/scraper"
	"radio-to-spotify/storage"
	"radio-to-spotify/utils"

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
	configHandler, err := utils.NewConfigHandler(stationFile)
	if err != nil {
		utils.Logger.Fatalf("Error loading config: %v", err)
	}

	store, err := storage.NewStorage(storageType, storagePath)
	if err != nil {
		utils.Logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		utils.Logger.Fatalf("Error initializing storage: %v", err)
	}

	stations, songs, err := scraper.FetchNowPlaying(configHandler, stationID)
	if err != nil {
		utils.Logger.Fatalf("Error fetching now playing: %v", err)
	}

	for i, station := range stations {
		if storeDryRun {
			utils.Logger.Infof("Dry run: would store song for station %s: %s - %s\n", station.ID, songs[i].Artist, songs[i].Title)
		} else {
			changed, err := store.StoreNowPlaying(station.ID, songs[i])
			if err != nil {
				utils.Logger.Fatalf("Error storing now playing for station %s: %v", station.ID, err)
			}
			if changed {
				utils.Logger.Infof("Stored song for station %s: %s - %s\n", station.ID, songs[i].Artist, songs[i].Title)

			} else {
				utils.Logger.Infof("Song hasn't changed for station %s: %s - %s\n", station.ID, songs[i].Artist, songs[i].Title)
			}
		}
	}
}
