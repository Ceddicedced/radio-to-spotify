package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"radio-to-spotify/scraper"
)

type FileStorage struct {
	mu    sync.Mutex
	songs map[string][]struct {
		scraper.Song
		Timestamp time.Time `json:"timestamp"`
	}
	filePath string
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return nil, err
	}

	fs := &FileStorage{
		songs: make(map[string][]struct {
			scraper.Song
			Timestamp time.Time `json:"timestamp"`
		}),
		filePath: filePath,
	}
	err := fs.loadFromFile()
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func (s *FileStorage) Init() error {
	return s.loadFromFile()
}

func (s *FileStorage) StoreNowPlaying(stationID string, song *scraper.Song) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSongs, exists := s.songs[stationID]
	if exists && len(lastSongs) > 0 {
		lastSong := lastSongs[len(lastSongs)-1]
		if lastSong.Artist == song.Artist && lastSong.Title == song.Title {
			return nil // Song hasn't changed
		}
	}

	// Append the new song to the list with timestamp
	songWithTimestamp := struct {
		scraper.Song
		Timestamp time.Time `json:"timestamp"`
	}{
		*song,
		time.Now(),
	}

	s.songs[stationID] = append(s.songs[stationID], songWithTimestamp)

	// Serialize the song list and save to file
	return s.saveToFile()
}

func (s *FileStorage) loadFromFile() error {
	dbFile := filepath.Join(s.filePath, "songs.json")
	file, err := os.Open(dbFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File does not exist, will create when storing
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&s.songs)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileStorage) saveToFile() error {
	dbFile := filepath.Join(s.filePath, "songs.json")
	file, err := os.Create(dbFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(s.songs)
}

func (s *FileStorage) GetNowPlaying(stationID string) (*scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSongs, exists := s.songs[stationID]
	if !exists || len(lastSongs) == 0 {
		return nil, errors.New("no song found for station")
	}

	return &lastSongs[len(lastSongs)-1].Song, nil
}

func (s *FileStorage) GetSongsSince(stationID string, sinceTime time.Time) ([]scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSongs, exists := s.songs[stationID]
	if !exists || len(lastSongs) == 0 {
		return nil, errors.New("no song found for station")
	}

	var songs []scraper.Song
	for _, song := range lastSongs {
		if song.Timestamp.After(sinceTime) {
			songs = append(songs, song.Song)
		}
	}

	return songs, nil
}

func (s *FileStorage) GetAllStations() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var stations []string
	for stationID := range s.songs {
		stations = append(stations, stationID)
	}
	return stations, nil
}
