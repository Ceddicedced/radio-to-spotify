package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"radio-to-spotify/scraper"
)

type FileStorage struct {
	mu    sync.Mutex
	songs map[string]*scraper.Song
	path  string
}

func NewFileStorage(path string) (*FileStorage, error) {
	return &FileStorage{
		songs: make(map[string]*scraper.Song),
		path:  path,
	}, nil
}

func (s *FileStorage) Init() error {
	files, err := os.ReadDir(s.path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			stationID := file.Name()
			stationID = stationID[:len(stationID)-len(".json")]
			song, err := s.loadLastSong(stationID)
			if err == nil {
				s.songs[stationID] = song
			}
		}
	}
	return nil
}

func (s *FileStorage) StoreNowPlaying(stationID string, song *scraper.Song) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSong, exists := s.songs[stationID]
	if exists && lastSong.Artist == song.Artist && lastSong.Title == song.Title {
		return nil // Song hasn't changed
	}

	s.songs[stationID] = song

	// Append the new song to the list with timestamp
	songWithTimestamp := struct {
		scraper.Song
		Timestamp time.Time `json:"timestamp"`
	}{
		*song,
		time.Now(),
	}

	// Load existing songs
	filePath := s.getFilePath(stationID)
	var songs []struct {
		scraper.Song
		Timestamp time.Time `json:"timestamp"`
	}

	file, err := os.Open(filePath)
	if err == nil {
		decoder := json.NewDecoder(file)
		decoder.Decode(&songs)
		file.Close()
	} else if !os.IsNotExist(err) {
		return err
	}

	songs = append(songs, songWithTimestamp)

	// Serialize the song list and save to file
	file, err = os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(songs); err != nil {
		return err
	}

	return nil
}

func (s *FileStorage) GetNowPlaying(stationID string) (*scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	song, exists := s.songs[stationID]
	if !exists {
		return nil, errors.New("no song found for station")
	}

	return song, nil
}

func (s *FileStorage) loadLastSong(stationID string) (*scraper.Song, error) {
	filePath := s.getFilePath(stationID)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var songs []struct {
		scraper.Song
		Timestamp time.Time `json:"timestamp"`
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&songs); err != nil {
		return nil, err
	}

	if len(songs) == 0 {
		return nil, errors.New("no song found")
	}

	return &songs[len(songs)-1].Song, nil
}

func (s *FileStorage) getFilePath(stationID string) string {
	return s.path + "/" + stationID + ".json"
}
