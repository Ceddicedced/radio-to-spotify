package spotify

import (
	"context"
	"fmt"
	"radio-to-spotify/config"
	"radio-to-spotify/scraper"
	"radio-to-spotify/storage"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
)

type SpotifyService struct {
	client        *spotify.Client
	configHandler *config.ConfigHandler
	store         storage.Storage
	logger        *logrus.Logger
}

func NewSpotifyService(logger *logrus.Logger, configHandler *config.ConfigHandler, store storage.Storage) (*SpotifyService, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		return nil, err
	}
	logger.Infof("Logged in as: %s", user.DisplayName)
	return &SpotifyService{
		client:        client,
		configHandler: configHandler,
		store:         store,
		logger:        logger,
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

	if station.PlaylistID == "" {
		return fmt.Errorf("no playlist ID found for station: %s", station.Name)
	}

	playlistID := spotify.ID(station.PlaylistID)

	err = s.ReplaceSongsInPlaylist(playlistID, songs)
	if err != nil {
		return err
	}

	s.logger.Debugf("Updated Spotify playlist for station: %s with time range: %s", station.Name, timeRange)
	return nil
}

func (s *SpotifyService) ReplaceSongsInPlaylist(playlistID spotify.ID, songs []scraper.Song) error {
	var trackIDs []spotify.ID
	for _, song := range songs {
		searchResults, err := s.client.Search(context.Background(), fmt.Sprintf("%s %s", song.Artist, song.Title), spotify.SearchTypeTrack)
		if err != nil {
			return err
		}
		if searchResults.Tracks.Total > 0 {
			trackIDs = append(trackIDs, searchResults.Tracks.Tracks[0].ID)
		}
	}

	err := s.client.ReplacePlaylistTracks(context.Background(), playlistID, trackIDs...)
	return err
}
