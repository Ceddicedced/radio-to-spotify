package cmd

import (
	"radio-to-spotify/spotify"
	"radio-to-spotify/storage"
	"radio-to-spotify/utils"
	"sync"

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

	spotifyService, err := spotify.NewSpotifyService( configHandler, store)
	utils.Logger.Infof("Updating Spotify playlist for range: %s", playlistRange)
	if err != nil {
		utils.Logger.Fatalf("Error initializing Spotify service: %v", err)
	}

	if stationID == "" {
		configStations := configHandler.GetAllStations()
		if len(configStations) == 0 {
			utils.Logger.Fatalf("No stations found in config")
		}
		var wg sync.WaitGroup
		wg.Add(len(configStations))
		for _, station := range configStations {
			go func(stationID string) {
				defer wg.Done()
				updateStation(spotifyService, stationID)
			}(station.ID)
		}
		wg.Wait()
	} else {
		updateStation(spotifyService, stationID)
	}

}

func updateStation(spotifyService *spotify.SpotifyService, stationID string) {
	utils.Logger.Infof("Updating Spotify playlist for station: %s", stationID)
	err := spotifyService.UpdateSpotifyPlaylist(stationID, playlistRange)
	if err != nil {
		utils.Logger.Errorf("Error updating Spotify playlist for station %s: %v", stationID, err)
	} else {
		utils.Logger.Infof("Updated Spotify playlist for station: %s", stationID)
	}
}
