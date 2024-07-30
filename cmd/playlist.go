package cmd

import (
	"radio-to-spotify/config"
	"radio-to-spotify/spotify"
	"radio-to-spotify/storage"

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

	store, err := storage.NewStorage(storageType, storagePath)
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	spotifyService, err := spotify.NewSpotifyService(logger)
	if err != nil {
		logger.Fatalf("Error initializing Spotify service: %v", err)
	}

	err = spotifyService.UpdateSpotifyPlaylist(configHandler, stationID, store)
	if err != nil {
		logger.Fatalf("Error updating Spotify playlist: %v", err)
	}
}
