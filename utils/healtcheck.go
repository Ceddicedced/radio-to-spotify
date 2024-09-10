package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HealthStatus struct {
	Status             string `json:"status"`
	Message            string `json:"message,omitempty"`
	SpotifyStatus      string `json:"spotify_status,omitempty"`
	StorageStatus      string `json:"storage_status,omitempty"`
	InternetStatus     string `json:"internet_status,omitempty"`
	LastFetchTime      string `json:"last_fetch_time,omitempty"`
	LastPlaylistUpdate string `json:"last_playlist_update,omitempty"`
}

var (
	lastFetchTime          time.Time
	fetchInterval          time.Duration
	lastPlaylistUpdate     time.Time
	playlistUpdateInterval time.Duration
)

// HealthChecker is an interface for checking the health of services like Spotify and storage
type HealthChecker interface {
	CheckHealth() (bool, string)
}

var (
	spotifyChecker HealthChecker
)

// SetLastUpdateTime updates the timestamp for a given type of update (fetch or playlist)
func SetLastUpdateTime(updateType string, t time.Time) {
	switch updateType {
	case "fetch":
		lastFetchTime = t
	case "playlist":
		lastPlaylistUpdate = t
	}
}

// HealthCheckHandler handles the health check requests
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	Logger.Debug("Health check request received")
	status := HealthStatus{
		Status:  "healthy",
		Message: "Service is running",
	}

	// Check Fetch Ticker
	if ok, message := checkInterval("fetchTicker", lastFetchTime, fetchInterval); !ok {
		status.Status = "unhealthy"
		status.Message = message
	}

	// Check Playlist Update Ticker
	if ok, message := checkInterval("playlistUpdateTicker", lastPlaylistUpdate, playlistUpdateInterval); !ok {
		status.Status = "unhealthy"
		status.Message = message
	}

	// Check Internet Connection
	if ok, message := checkInternetConnection(); !ok {
		status.Status = "unhealthy"
		status.Message = message
	}

	// Check Spotify Connection
	if spotifyChecker != nil {
		if ok, message := spotifyChecker.CheckHealth(); !ok {
			status.Status = "unhealthy"
			status.SpotifyStatus = message
		}
	}

	status.LastFetchTime = lastFetchTime.Format(time.RFC3339)
	status.LastPlaylistUpdate = lastPlaylistUpdate.Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// StartHealthCheckServer starts an HTTP server to serve health checks
func StartHealthCheckServer(port int, fetchIntv, playlistIntv time.Duration, spotify HealthChecker) {
	fetchInterval = fetchIntv
	playlistUpdateInterval = playlistIntv
	spotifyChecker = spotify

	SetLastUpdateTime("fetch", time.Now())
	SetLastUpdateTime("playlist", time.Now()) // Initialize the playlist update time

	http.HandleFunc("/health", HealthCheckHandler)
	addr := fmt.Sprintf(":%d", port)
	Logger.Infof("Starting health check server on port %d", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		Logger.Fatalf("Error starting health check server: %v", err)
	}
}

// checkInterval is a generic function to check if a ticker (like fetch or playlist) is running within the expected interval
func checkInterval(tickerName string, lastUpdateTime time.Time, expectedInterval time.Duration) (bool, string) {
	timeSinceLastUpdate := time.Since(lastUpdateTime)
	delayDuration := timeSinceLastUpdate - expectedInterval
	if expectedInterval == 0 {
		return true, fmt.Sprintf("%s is not expected to run", tickerName)
	}
	if delayDuration.Round(time.Second) > 30*time.Second {
		return false, fmt.Sprintf("%s is delayed by %s", tickerName, delayDuration.Round(time.Second))
	}
	return true, fmt.Sprintf("%s is running as expected", tickerName)
}

// Checks if internet connection is available
func checkInternetConnection() (bool, string) {
	_, err := http.Get("https://go.dev/")
	if err != nil {
		return false, "Internet connection is down"
	}
	return true, "Internet connection is up"
}
