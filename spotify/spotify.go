package spotify

import (
	"fmt"
	"radio-to-spotify/storage"
)

func CreateSpotifyPlaylist(stationID string, store storage.Storage) error {
	song, err := store.GetNowPlaying(stationID)
	if err != nil {
		return fmt.Errorf("error retrieving now playing song: %v", err)
	}

	// Simulate creating a Spotify playlist
	// You would need to use the Spotify API here
	fmt.Printf("Creating Spotify playlist with the song: %s - %s\n", song.Artist, song.Title)
	return nil
}
