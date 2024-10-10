package storage

import (
	"container/list"
	"context"
	"fmt"
	"radio-to-spotify/utils"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type CacheItem struct {
	TrackID string
}

type SongCache struct {
	mu         sync.Mutex
	cache      map[string]*list.Element // Map from "artist - title" to *list.Element for constant-time lookups
	orderList  *list.List               // Doubly linked list to maintain LRU order
	redis      *redis.Client            // Redis client (optional)
	ctx        context.Context          // Context for Redis operations
	expiration time.Duration            // Expiration time for cached items
	maxSize    int                      // Maximum cache size before eviction
}

// List item containing key and value
type entry struct {
	key   string
	value CacheItem
}

// Initialize the cache with optional Redis, expiration, and size limit
func NewSongCache() *SongCache {
	expiration := getExpiration()
	maxSize := getMaxSize()
	redisClient := initializeRedis()

	if redisClient == nil {
		utils.Logger.Debug("Redis is not configured, using in-memory cache only")
	} else {
		utils.Logger.Infof("Redis configured, caching will use Redis with fallback to in-memory")
	}

	return &SongCache{
		cache:      make(map[string]*list.Element),
		orderList:  list.New(),
		redis:      redisClient,
		ctx:        context.Background(),
		expiration: expiration,
		maxSize:    maxSize,
	}
}

// Add a song to the cache
func (sc *SongCache) AddToCache(artist, title, trackID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	key := fmt.Sprintf("%s - %s", artist, title)

	// If item exists in cache, move it to the front (most recently used)
	if element, found := sc.cache[key]; found {
		sc.orderList.MoveToFront(element)
		element.Value.(*entry).value.TrackID = trackID
		utils.Logger.Debugf("Updated song in cache: %s - %s -> %s", artist, title, trackID)
		return
	}

	// Add new item to cache and list
	newEntry := &entry{
		key:   key,
		value: CacheItem{TrackID: trackID},
	}
	element := sc.orderList.PushFront(newEntry)
	sc.cache[key] = element

	utils.Logger.Debugf("Added song to in-memory cache: %s - %s -> %s", artist, title, trackID)

	// Evict oldest if cache is full
	if sc.orderList.Len() > sc.maxSize {
		sc.evictOldest()
	}

	// Add to Redis with expiration if Redis is configured
	if sc.redis != nil {
		err := sc.redis.Set(sc.ctx, key, trackID, sc.expiration).Err()
		if err != nil {
			utils.Logger.Warnf("Failed to add song to Redis: %v", err)
		} else {
			utils.Logger.Debugf("Added song to Redis: %s - %s -> %s", artist, title, trackID)
		}
	}
}

// Get a song from the cache
func (sc *SongCache) GetFromCache(artist, title string) (string, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	key := fmt.Sprintf("%s - %s", artist, title)

	// Check in-memory cache
	if element, found := sc.cache[key]; found {
		// Move accessed item to the front (most recently used)
		sc.orderList.MoveToFront(element)
		utils.Logger.Debugf("Got song from in-memory cache: %s - %s -> %s", artist, title, element.Value.(*entry).value.TrackID)
		return element.Value.(*entry).value.TrackID, true
	}

	// Check Redis if not found in-memory
	if sc.redis != nil {
		trackID, err := sc.redis.Get(sc.ctx, key).Result()
		if err == redis.Nil {
			utils.Logger.Debugf("Song not found in Redis cache: %s - %s", artist, title)
			return "", false
		} else if err != nil {
			utils.Logger.Warnf("Failed to retrieve song from Redis: %v", err)
			return "", false
		}

		// Add Redis entry to in-memory cache
		newEntry := &entry{
			key:   key,
			value: CacheItem{TrackID: trackID},
		}
		element := sc.orderList.PushFront(newEntry)
		sc.cache[key] = element

		utils.Logger.Debugf("Got song from Redis and added to in-memory cache: %s - %s -> %s", artist, title, trackID)
		return trackID, true
	}

	utils.Logger.Debugf("Song not found in any cache: %s - %s", artist, title)
	return "", false
}

// Evict the oldest entry from the in-memory cache
func (sc *SongCache) evictOldest() {
	oldest := sc.orderList.Back()
	if oldest != nil {
		entry := oldest.Value.(*entry)
		delete(sc.cache, entry.key)
		sc.orderList.Remove(oldest)
		utils.Logger.Debugf("Evicted oldest item from in-memory cache: %s", entry.key)
	}
}

// Get expiration duration from environment variable or default value
func getExpiration() time.Duration {
	expirationStr := utils.GetEnv("CACHE_EXPIRATION", "756h") // 4 weeks
	expiration, err := time.ParseDuration(expirationStr)
	if err != nil {
		utils.Logger.Warnf("Invalid CACHE_EXPIRATION value: %v. Using default 4w.", err)
		return 756 * time.Hour
	}
	return expiration
}

// Get max size for in-memory cache from environment variable or default value
func getMaxSize() int {
	maxSizeStr := utils.GetEnv("CACHE_MAX_SIZE", "10000")
	maxSize, err := strconv.Atoi(maxSizeStr)
	if err != nil {
		utils.Logger.Warnf("Invalid CACHE_MAX_SIZE value: %v. Using default 1000.", err)
		return 1000
	}
	return maxSize
}

// Initialize Redis client via environment variables, with support for redis:// connection strings
func initializeRedis() *redis.Client {
	redisURL := utils.GetEnv("REDIS_URL", "")
	if redisURL == "" {
		return nil // No Redis URL provided, return nil to indicate no Redis configuration
	}

	// Parse the redis:// connection string
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		utils.Logger.Warnf("Invalid Redis URL: %v. Skipping Redis initialization.", err)
		return nil
	}

	redisClient := redis.NewClient(opt)

	// Ping Redis to ensure connection is working
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		utils.Logger.Warnf("Failed to connect to Redis: %v", err)
		return nil
	}

	utils.Logger.Infof("Successfully connected to Redis at %s", opt.Addr)
	return redisClient
}
