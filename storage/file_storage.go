package storage

import (
	"encoding/json"
	"errors"
	"os"

	"radio-to-spotify/scraper"
)

type FileStorage struct {
	BaseStorage
	path string
}

func NewFileStorage(path string) (*FileStorage, error) {
	return &FileStorage{
		BaseStorage: BaseStorage{
			songs: make(map[string]*scraper.Song),
		},
		path: path,
	}, nil
}

func (s *FileStorage) StoreNowPlaying(stationID string, song *scraper.Song) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the song has changed
	if existingSong, exists := s.songs[stationID]; exists {
		if existingSong.Artist == song.Artist && existingSong.Title == song.Title {
			return nil // Song hasn't changed
		}
	}

	// Update the in-memory store
	s.songs[stationID] = song

	// Serialize the song and save to file
	filePath := s.getFilePath(stationID)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(song); err != nil {
		return err
	}

	return nil
}

func (s *FileStorage) GetNowPlaying(stationID string) (*scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the song is in the in-memory store
	if song, exists := s.songs[stationID]; exists {
		return song, nil
	}

	// Load the song from the file
	filePath := s.getFilePath(stationID)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("no song found for station")
		}
		return nil, err
	}
	defer file.Close()

	var song scraper.Song
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&song); err != nil {
		return nil, err
	}

	// Update the in-memory store
	s.songs[stationID] = &song

	return &song, nil
}

func (s *FileStorage) getFilePath(stationID string) string {
	return s.path + "/" + stationID + ".json"
}
