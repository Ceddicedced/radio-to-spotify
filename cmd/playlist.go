package cmd

import (
	"github.com/spf13/cobra"

	"radio-to-spotify/spotify"
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
	err := spotify.CreateSpotifyPlaylist(stationID, store)
	if err != nil {
		logger.Fatalf("Error creating Spotify playlist: %v", err)
	}
}
