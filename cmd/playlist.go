package cmd

import (
	"radio-to-spotify/config"
	"radio-to-spotify/spotify"

	"github.com/spf13/cobra"
)

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Create Spotify playlist for the last hour of songs",
	Run: func(cmd *cobra.Command, args []string) {
		executePlaylist()
	},
}

func init() {
	rootCmd.AddCommand(playlistCmd)
}

func executePlaylist() {
	configHandler, err := config.NewConfigHandler(stationFile)
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	err = spotify.CreateSpotifyPlaylist(configHandler, stationID, store)
	if err != nil {
		logger.Fatalf("Error creating Spotify playlist: %v", err)
	}
}
