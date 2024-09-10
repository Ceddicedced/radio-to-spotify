package cmd

import (
	"radio-to-spotify/scraper"
	"radio-to-spotify/utils"

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
	configHandler, err := utils.NewConfigHandler(stationFile)
	if err != nil {
		utils.Logger.Fatalf("Error loading config: %v", err)
	}

	stations, songs, err := scraper.FetchNowPlaying(configHandler, stationID)
	if err != nil {
		utils.Logger.Fatalf("Error fetching now playing: %v", err)
	}

	for i, station := range stations {
		utils.Logger.Infof("Station: %s, Song: %s - %s", station.Name, songs[i].Artist, songs[i].Title)
	}
}
