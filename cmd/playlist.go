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
	playlistCmd.Flags().StringVar(&playlistRange, "playlist-range", "lastday", "Time range for playlist update (lasthour, lastday, lastweek)")
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

	spotifyService, err := spotify.NewSpotifyService(logger, configHandler, store)
	if err != nil {
		logger.Fatalf("Error initializing Spotify service: %v", err)
	}

	if stationID == "" {
		configStations := configHandler.GetAllStations()
		if len(configStations) == 0 {
			logger.Fatalf("No stations found in config")
		}
		for _, station := range configStations {
			updateStation(spotifyService, station.ID)
		}
	} else {
		updateStation(spotifyService, stationID)
	}

}

func updateStation(spotifyService *spotify.SpotifyService, stationID string) {
	err := spotifyService.UpdateSpotifyPlaylist(stationID, playlistRange)
	if err != nil {
		logger.Errorf("Error updating Spotify playlist for station %s: %v", stationID, err)
	} else {
		logger.Infof("Updated Spotify playlist for station: %s", stationID)
	}
}
