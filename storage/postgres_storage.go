package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"radio-to-spotify/scraper"

	_ "github.com/lib/pq"
)

type PostgreSQLStorage struct {
	mu    sync.Mutex
	songs map[string]*scraper.Song
	db    *sql.DB
}

func NewPostgreSQLStorage(connStr string) (*PostgreSQLStorage, error) {
	if connStr == "data" {
		return nil, errors.New("missing connection string for PostgreSQL storage. Example: postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full | provide via --storage-path")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	storage := &PostgreSQLStorage{
		songs: make(map[string]*scraper.Song),
		db:    db,
	}

	err = storage.initPostgreSQL()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *PostgreSQLStorage) initPostgreSQL() error {
	return nil
}

func (s *PostgreSQLStorage) Init() error {
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

func (s *PostgreSQLStorage) getStationTables() ([]string, error) {
	var tables []string
	rows, err := s.db.Query(`SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'station_%'`)
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

func (s *PostgreSQLStorage) loadLastSong(stationID string) (*scraper.Song, error) {
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

func (s *PostgreSQLStorage) createStationTable(stationID string) error {
	_, err := s.db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS station_%s (
		id SERIAL PRIMARY KEY,
		artist TEXT,
		title TEXT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`, stationID))
	return err
}

func (s *PostgreSQLStorage) StoreNowPlaying(stationID string, song *scraper.Song) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSong, exists := s.songs[stationID]
	if exists && lastSong.Artist == song.Artist && lastSong.Title == song.Title {
		return false, nil // Song hasn't changed
	}

	s.songs[stationID] = song

	// Ensure the table for the station exists
	err := s.createStationTable(stationID)
	if err != nil {
		return false, err
	}

	// Insert the new song into the station-specific table
	_, err = s.db.Exec(fmt.Sprintf(`INSERT INTO station_%s (artist, title) VALUES ($1, $2)`, stationID),
		song.Artist, song.Title)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *PostgreSQLStorage) GetNowPlaying(stationID string) (*scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	song, exists := s.songs[stationID]
	if !exists {
		return nil, errors.New("no song found for station")
	}

	return song, nil
}

func (s *PostgreSQLStorage) GetSongsSince(stationID string, sinceTime time.Time) ([]scraper.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.songs[stationID]; !exists {
		return nil, errors.New("no song found for station")
	}
	rows, err := s.db.Query(fmt.Sprintf(`SELECT artist, title FROM station_%s WHERE timestamp > $1`, stationID), sinceTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []scraper.Song
	for rows.Next() {
		var song scraper.Song
		if err := rows.Scan(&song.Artist, &song.Title); err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, rows.Err()
}

func (s *PostgreSQLStorage) GetAllStations() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tables, err := s.getStationTables()
	if err != nil {
		return nil, err
	}

	var stations []string
	for _, table := range tables {
		stationID := strings.TrimPrefix(table, "station_")
		stations = append(stations, stationID)
	}

	return stations, nil
}
