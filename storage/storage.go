package storage

import (
	"errors"
	"sync"

	"radio-to-spotify/scraper"
)

type Storage struct {
	mu    sync.Mutex
	songs map[string]*scraper.Song
}

func NewStorage() *Storage {
	return &Storage{
		songs: make(map[string]*scraper.Song),
	}
}

func (s *Storage) StoreNowPlaying(stationID string, song *scraper.Song) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingSong, exists := s.songs[stationID]; exists {
		if existingSong.Artist == song.Artist && existingSong.Title == song.Title {
			return nil // Song hasn't changed
		}
	}
	s.songs[stationID] = song
	return nil
}

func (s *Storage) GetNowPlaying(stationID string) (*scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	song, exists := s.songs[stationID]
	if !exists {
		return nil, errors.New("no song found for station")
	}
	return song, nil
}
