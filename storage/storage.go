package storage

import (
	"fmt"
	"time"

	"radio-to-spotify/scraper"
)

type Storage interface {
	StoreNowPlaying(stationID string, song *scraper.Song) error
	GetNowPlaying(stationID string) (*scraper.Song, error)
	GetSongsSince(stationID string, sinceTime time.Time) ([]scraper.Song, error)
	GetAllStations() ([]string, error)
	Init() error
}

func NewStorage(storageType, storagePath string) (Storage, error) {
	switch storageType {
	case "file":
		return NewFileStorage(storagePath)
	case "sqlite":
		return NewSQLiteStorage(storagePath)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
