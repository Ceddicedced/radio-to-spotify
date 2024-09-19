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

// var (
// 	spotifyChecker HealthChecker
// )

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
		Logger.Warnf("Fetch ticker is delayed by %s", message)
		status.Status = "unhealthy"
		status.Message = message
	}
	Logger.Debugf("Last fetch time: %v", lastFetchTime)

	// Check Playlist Update Ticker
	if ok, message := checkInterval("playlistUpdateTicker", lastPlaylistUpdate, playlistUpdateInterval); !ok {
		Logger.Warnf("Playlist update ticker is delayed by %s", message)
		status.Status = "unhealthy"
		status.Message = message
	}
	Logger.Debugf("Last playlist update time: %v", lastPlaylistUpdate)

	// Check Internet Connection
	if ok, message := checkInternetConnection(); !ok {
		Logger.Warnf("Internet connection is down")
		status.Status = "unhealthy"
		status.Message = message
	} else {
		Logger.Debugf("Internet connection is up")
	}

	// // Check Spotify Connection
	// if spotifyChecker != nil {
	// 	if ok, message := spotifyChecker.CheckHealth(); !ok {
	// 		Logger.Warnf("Spotify connection is down")
	// 		status.Status = "unhealthy"
	// 		status.Message = message
	// 	} else {
	// 		Logger.Debugf("Spotify connection is up")
	// 	}
	// }

	status.LastFetchTime = lastFetchTime.Format(time.RFC3339)
	status.LastPlaylistUpdate = lastPlaylistUpdate.Format(time.RFC3339)

	Logger.Debugf("Health check response: %v", status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// StartHealthCheckServer starts an HTTP server to serve health checks
func StartHealthCheckServer(port int, fetchIntv, playlistIntv time.Duration, spotify HealthChecker) {
	// Initialize the global variables
	fetchInterval = fetchIntv
	playlistUpdateInterval = playlistIntv
	// spotifyChecker = spotify

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
	Logger.Debugf("Time since last %s: %s", tickerName, timeSinceLastUpdate)
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
