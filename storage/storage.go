package storage

import (
	"fmt"
	"sync"

	"radio-to-spotify/scraper"
)

type Storage interface {
	StoreNowPlaying(stationID string, song *scraper.Song) error
	GetNowPlaying(stationID string) (*scraper.Song, error)
}

type BaseStorage struct {
	mu    sync.Mutex
	songs map[string]*scraper.Song
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
