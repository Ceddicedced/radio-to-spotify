package storage

import (
	"database/sql"
	"errors"

	"radio-to-spotify/scraper"
)

type SQLiteStorage struct {
	BaseStorage
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &SQLiteStorage{
		BaseStorage: BaseStorage{
			songs: make(map[string]*scraper.Song),
		},
		db: db,
	}

	err = storage.initSQLite()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *SQLiteStorage) initSQLite() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS now_playing (
		station_id TEXT PRIMARY KEY,
		artist TEXT,
		title TEXT
	)`)
	return err
}

func (s *SQLiteStorage) StoreNowPlaying(stationID string, song *scraper.Song) error {
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

	// Store in SQLite database
	_, err := s.db.Exec(`INSERT OR REPLACE INTO now_playing (station_id, artist, title) VALUES (?, ?, ?)`,
		stationID, song.Artist, song.Title)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLiteStorage) GetNowPlaying(stationID string) (*scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the song is in the in-memory store
	if song, exists := s.songs[stationID]; exists {
		return song, nil
	}

	// Load from SQLite database
	row := s.db.QueryRow(`SELECT artist, title FROM now_playing WHERE station_id = ?`, stationID)
	var song scraper.Song
	err := row.Scan(&song.Artist, &song.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no song found for station")
		}
		return nil, err
	}

	// Update the in-memory store
	s.songs[stationID] = &song
	return &song, nil
}
