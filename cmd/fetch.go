package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"radio-to-spotify/scraper"
)

func init() {
	rootCmd.AddCommand(fetchCmd)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch now playing songs from radio stations",
	Run: func(cmd *cobra.Command, args []string) {
		executeFetch()
	},
}

func executeFetch() {
	stationSongs, err := scraper.FetchNowPlaying(stationFile, logger, stationID)
	if err != nil {
		logger.Fatalf("Error fetching now playing: %v", err)
	}

	for _, stationSong := range stationSongs {
		fmt.Printf("Station: %s (%s) - Artist: %s, Title: %s\n", stationSong.Station, stationSong.StationID, stationSong.Song.Artist, stationSong.Song.Title)
	}
}
