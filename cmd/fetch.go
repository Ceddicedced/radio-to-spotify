package cmd

import (
	"radio-to-spotify/config"
	"radio-to-spotify/scraper"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fetchCmd)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch now playing songs for radio stations",
	Run: func(cmd *cobra.Command, args []string) {
		executeFetch()
	},
}

func executeFetch() {
	configHandler, err := config.NewConfigHandler(stationFile)
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	stations, songs, err := scraper.FetchNowPlaying(configHandler, logger, stationID)
	if err != nil {
		logger.Fatalf("Error fetching now playing: %v", err)
	}

	for i, station := range stations {
		logger.Infof("Station: %s, Song: %s - %s", station.Name, songs[i].Artist, songs[i].Title)
	}
}
