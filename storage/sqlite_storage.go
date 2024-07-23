package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"radio-to-spotify/scraper"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	mu    sync.Mutex
	songs map[string]*scraper.Song
	db    *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
		return nil, err
	}

	// Join the directory path with the database file name
	dbFile := filepath.Join(dbPath, "db.sqlite")

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	storage := &SQLiteStorage{
		songs: make(map[string]*scraper.Song),
		db:    db,
	}

	err = storage.initSQLite()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *SQLiteStorage) initSQLite() error {
	// No need to create a general now_playing table since each station will have its own table
	return nil
}

func (s *SQLiteStorage) Init() error {
	tables, err := s.getStationTables()
	if err != nil {
		return err
	}

	for _, table := range tables {
		stationID := strings.TrimPrefix(table, "station_")
		song, err := s.loadLastSong(stationID)
		if err == nil {
			s.songs[stationID] = song
		}
	}
	return nil
}

func (s *SQLiteStorage) getStationTables() ([]string, error) {
	var tables []string
	rows, err := s.db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'station_%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

func (s *SQLiteStorage) loadLastSong(stationID string) (*scraper.Song, error) {
	row := s.db.QueryRow(fmt.Sprintf(`SELECT artist, title FROM station_%s ORDER BY timestamp DESC LIMIT 1`, stationID))
	var song scraper.Song
	err := row.Scan(&song.Artist, &song.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no song found for station")
		}
		return nil, err
	}
	return &song, nil
}

func (s *SQLiteStorage) createStationTable(stationID string) error {
	_, err := s.db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS station_%s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		artist TEXT,
		title TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`, stationID))
	return err
}

func (s *SQLiteStorage) StoreNowPlaying(stationID string, song *scraper.Song) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSong, exists := s.songs[stationID]
	if exists && lastSong.Artist == song.Artist && lastSong.Title == song.Title {
		return nil // Song hasn't changed
	}

	s.songs[stationID] = song

	// Ensure the table for the station exists
	err := s.createStationTable(stationID)
	if err != nil {
		return err
	}

	// Insert the new song into the station-specific table
	_, err = s.db.Exec(fmt.Sprintf(`INSERT INTO station_%s (artist, title) VALUES (?, ?)`, stationID),
		song.Artist, song.Title)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLiteStorage) GetNowPlaying(stationID string) (*scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	song, exists := s.songs[stationID]
	if !exists {
		return nil, errors.New("no song found for station")
	}

	return song, nil
}
