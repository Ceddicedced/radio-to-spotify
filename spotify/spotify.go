package spotify

import (
	"context"
	"fmt"
	"radio-to-spotify/scraper"
	"radio-to-spotify/storage"
	"radio-to-spotify/utils"
	"time"

	"github.com/zmb3/spotify/v2"
)

type SpotifyService struct {
	client        *spotify.Client
	configHandler *utils.ConfigHandler
	store         storage.Storage
	cache         *storage.SongCache
}

func NewSpotifyService(configHandler *utils.ConfigHandler, store storage.Storage) (*SpotifyService, error) {
	utils.Logger.Debug("Initializing Spotify service")
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	utils.Logger.Debug("Got Spotify client")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	user, err := client.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	utils.Logger.Infof("Logged in as: %s", user.DisplayName)

	cache := storage.NewSongCache()

	return &SpotifyService{
		client:        client,
		configHandler: configHandler,
		store:         store,
		cache:         cache,
	}, nil
}

func (s *SpotifyService) UpdateSpotifyPlaylist(stationID, timeRange string) error {
	_, err := s.client.CurrentUser(context.Background())
	if err != nil {
		return err
	}

	station, err := s.configHandler.GetStationByID(stationID)
	if err != nil {
		return err
	}

	var songs []scraper.Song
	switch timeRange {
	case "lasthour":
		songs, err = s.store.GetSongsSince(stationID, time.Now().Add(-1*time.Hour))
	case "lastday":
		songs, err = s.store.GetSongsSince(stationID, time.Now().Add(-24*time.Hour))
	case "lastweek":
		songs, err = s.store.GetSongsSince(stationID, time.Now().Add(-7*24*time.Hour))
	default:
		return fmt.Errorf("invalid time range: %s", timeRange)
	}
	if err != nil {
		return err
	}
	utils.Logger.Debugf("Updating Spotify Playlist with %d songs for station: %s with time range: %s", len(songs), station.Name, timeRange)

	if station.PlaylistID == "" {
		return fmt.Errorf("no playlist ID found for station: %s", station.Name)
	}

	playlistID := spotify.ID(station.PlaylistID)

	err = s.ReplaceSongsInPlaylist(playlistID, songs)
	if err != nil {
		return err
	}

	utils.Logger.Debugf("Updated Spotify playlist for station: %s with time range: %s", station.Name, timeRange)
	return nil
}

func (s *SpotifyService) ReplaceSongsInPlaylist(playlistID spotify.ID, songs []scraper.Song) error {
	var trackIDs []spotify.ID

	for _, song := range songs {
		// Check if the song is already in the cache
		if cachedID, found := s.cache.GetFromCache(song.Artist, song.Title); found {
			trackID := spotify.ID(cachedID)
			trackIDs = append(trackIDs, trackID)
			utils.Logger.Debugf("Using cached track ID for: %s - %s", song.Artist, song.Title)
			continue
		}

		searchResults, err := s.client.Search(context.Background(), fmt.Sprintf("%s %s", song.Artist, song.Title), spotify.SearchTypeTrack)
		if err != nil {
			utils.Logger.Warnf("Error searching for track: %s by %s: %v", song.Title, song.Artist, err)
			return err
		}
		if searchResults.Tracks.Total > 0 && len(searchResults.Tracks.Tracks) > 0 {
			track := searchResults.Tracks.Tracks[0]
			utils.Logger.Debugf("Found track: %s - %s", track.Artists[0].Name, track.Name)
			trackIDs = append(trackIDs, track.ID)
			s.cache.AddToCache(song.Artist, song.Title, track.ID.String())
		}
		if searchResults.Tracks.Total == 0 {
			utils.Logger.Warnf("No track found for: %s - %s", song.Artist, song.Title)
		}
		if len(searchResults.Tracks.Tracks) == 0 {
			utils.Logger.Warnf("Track Page is empty for: %s - %s (%s)", song.Artist, song.Title, searchResults.Tracks.Endpoint)
		}

	}

	utils.Logger.Debugf("Replacing playlist %s with %d tracks", playlistID, len(trackIDs))

	// Replace the entire playlist with the new tracks
	if len(trackIDs) > 100 {
		err := s.replacePlaylistTracksInBatches(playlistID, trackIDs)
		if err != nil {
			return err
		}
	} else {
		err := s.client.ReplacePlaylistTracks(context.Background(), playlistID, trackIDs...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SpotifyService) replacePlaylistTracksInBatches(playlistID spotify.ID, trackIDs []spotify.ID) error {
	// Clear the playlist first
	err := s.client.ReplacePlaylistTracks(context.Background(), playlistID)
	if err != nil {
		return err
	}

	for i := 0; i < len(trackIDs); i += 100 {
		end := i + 100
		if end > len(trackIDs) {
			end = len(trackIDs)
		}
		batch := trackIDs[i:end]
		_, err := s.client.AddTracksToPlaylist(context.Background(), playlistID, batch...)
		if err != nil {
			utils.Logger.Errorf("Error adding batch of tracks to playlist %s: %v", playlistID, err)
			return err
		}
		utils.Logger.Debugf("Added batch of %d tracks to playlist %s", len(batch), playlistID)
	}

	return nil
}
