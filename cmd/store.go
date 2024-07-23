package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"radio-to-spotify/scraper"
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
	stationSongs, err := scraper.FetchNowPlaying(stationFile, logger, stationID)
	if err != nil {
		logger.Fatalf("Error fetching now playing: %v", err)
	}

	for _, stationSong := range stationSongs {
		if storeDryRun {
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
