package storage

import (
	"fmt"
	"radio-to-spotify/utils"
	"sync"
)

type SongCache struct {
	mu    sync.Mutex
	cache map[string]string // A map from "artist - title" to Spotify track ID
}

// Initialize the cache
func NewSongCache() *SongCache {
	utils.Logger.Debug("Initializing song cache")
	return &SongCache{
		cache: make(map[string]string),
	}
}

// Add a song to the cache
func (sc *SongCache) AddToCache(artist, title string, trackID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	key := fmt.Sprintf("%s - %s", artist, title)
	sc.cache[key] = trackID

	utils.Logger.Debugf("Added song to cache: %s - %s -> %s", artist, title, trackID)
}

// Check if a song is in the cache
func (sc *SongCache) GetFromCache(artist, title string) (string, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	key := fmt.Sprintf("%s - %s", artist, title)
	id, found := sc.cache[key]

	utils.Logger.Debugf("Got song from cache: %s - %s -> %s", artist, title, id)

	return id, found
}
