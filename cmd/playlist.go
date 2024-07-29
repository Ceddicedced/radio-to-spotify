package cmd

import (
	"radio-to-spotify/spotify"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(playlistCmd)
}

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Create Spotify playlist for the last hour of songs",
	Run: func(cmd *cobra.Command, args []string) {
		executePlaylist()
	},
}

func executePlaylist() {
	if stationID == "" {
		logger.Fatalf("Station ID is required")
	}

	err := spotify.CreateSpotifyPlaylist(stationID, store)
	if err != nil {
		logger.Fatalf("Error creating Spotify playlist: %v", err)
	}
}
