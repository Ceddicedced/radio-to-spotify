package storage

import (
	"database/sql"
	"errors"
	"fmt"
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
	fmt.Println("Creating SQLite storage")
	db, err := sql.Open("sqlite3", dbPath+"db.sqlite")
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
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS now_playing (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		station_id TEXT,
		artist TEXT,
		title TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

func (s *SQLiteStorage) Init() error {
	rows, err := s.db.Query(`SELECT station_id, artist, title FROM now_playing WHERE timestamp = (SELECT MAX(timestamp) FROM now_playing AS np WHERE np.station_id = now_playing.station_id)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var stationID, artist, title string
		if err := rows.Scan(&stationID, &artist, &title); err != nil {
			return err
		}
		s.songs[stationID] = &scraper.Song{Artist: artist, Title: title}
	}

	return rows.Err()
}

func (s *SQLiteStorage) StoreNowPlaying(stationID string, song *scraper.Song) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSong, exists := s.songs[stationID]
	if exists && lastSong.Artist == song.Artist && lastSong.Title == song.Title {
		return nil // Song hasn't changed
	}

	s.songs[stationID] = song

	// Insert the new song into the database
	_, err := s.db.Exec(`INSERT INTO now_playing (station_id, artist, title) VALUES (?, ?, ?)`,
		stationID, song.Artist, song.Title)
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
