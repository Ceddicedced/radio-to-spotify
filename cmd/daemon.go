package cmd

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"radio-to-spotify/scraper"
	"radio-to-spotify/spotify"
	"radio-to-spotify/storage"
	"radio-to-spotify/utils"

	"github.com/spf13/cobra"
)

var (
	noStore                  bool
	noPlaylist               bool
	fetchInterval            time.Duration
	playlistUpdateInterval   time.Duration
	playlistRange            string
	sessionKeepAliveInterval time.Duration
)

type ScraperService struct {
	FetchInterval            time.Duration
	PlaylistUpdateInterval   time.Duration
	SessionKeepAliveInterval time.Duration
	stopScraper              chan struct{}
	configHandler            *utils.ConfigHandler
	storage                  storage.Storage
	spotify                  *spotify.SpotifyService
}

func (s *ScraperService) Start() {
	utils.Logger.Infof("Starting scraper service with fetch interval %v", s.FetchInterval)
	fetchTicker := time.NewTicker(s.FetchInterval)
	var playlistUpdateTicker *time.Ticker

	if !noPlaylist {
		utils.Logger.Infof("Starting playlist update ticker with interval %v", s.PlaylistUpdateInterval)
		playlistUpdateTicker = time.NewTicker(s.PlaylistUpdateInterval)
	} else { // If no playlist update, don't start the tickers / Kinda hacky
		utils.Logger.Info("Running without playlist update")
		playlistUpdateTicker = time.NewTicker(1)
		playlistUpdateTicker.Stop()
	}

	wg := sync.WaitGroup{}
	for {
		select {
		case <-fetchTicker.C:
			utils.Logger.Debug("FetchTicker tick")
			wg.Add(1)
			go s.fetchNowPlaying(&wg)
		case <-playlistUpdateTicker.C:
			utils.Logger.Debug("PlaylistUpdateTicker tick")
			wg.Add(1)
			go s.updatePlaylists(&wg)
		case <-s.stopScraper:
			utils.Logger.Debug("Stop signal received")
			utils.Logger.Info("Waiting for goroutines to finish")
			wg.Wait()
			fetchTicker.Stop()
			if playlistUpdateTicker != nil {
				playlistUpdateTicker.Stop()
			}
			return
		}
	}
}

func (s *ScraperService) Stop() {
	utils.Logger.Info("Stopping scraper service")
	close(s.stopScraper)
}

func (s *ScraperService) fetchNowPlaying(wg *sync.WaitGroup) {
	defer wg.Done()
	utils.SetLastUpdateTime("fetch", time.Now())
	utils.Logger.Debugf("Fetching now playing songs")
	var storedCount, songCount int

	if !noStore {
		stations, songs, err := scraper.FetchNowPlaying(s.configHandler, stationID)
		if err != nil {
			utils.Logger.Warnf("Error fetching now playing: %v", err)
			return
		}
		songCount = len(songs)

		for i, station := range stations {
			changed, err := s.storage.StoreNowPlaying(station.ID, songs[i])
			if err != nil {
				utils.Logger.Errorf("Error storing now playing for station %s: %v", station.ID, err)
			} else {
				if changed {
					// Only count stored songs that have changed
					storedCount++
					utils.Logger.Debugf("Stored song for station %s: %s - %s", station.ID, songs[i].Artist, songs[i].Title)
				} else {
					utils.Logger.Debugf("No change in song for station %s: %s - %s", station.ID, songs[i].Artist, songs[i].Title)
				}
			}
		}
	}

	utils.Logger.Infof("Fetched %d stations, stored %d songs", songCount, storedCount)
}

func (s *ScraperService) updatePlaylists(wg *sync.WaitGroup) {
	defer wg.Done()
	if noPlaylist {
		return
	}

	utils.SetLastUpdateTime("playlist", time.Now())
	utils.Logger.Debugf("Updating playlists")
	var playlistCount int

	stations, err := s.storage.GetAllStations()
	if err != nil {
		utils.Logger.Errorf("Error getting all stations: %v", err)
		return
	}
	if stationID != "" {
		stations = []string{stationID}
	} else {
		utils.Logger.Debugf("Updating playlists for all stations")
	}

	for _, stationID := range stations {
		err := s.spotify.UpdateSpotifyPlaylist(stationID, playlistRange)
		if err != nil {
			utils.Logger.Errorf("Error updating Spotify playlist for station %s: %v", stationID, err)
		} else {
			playlistCount++
		}
	}

	utils.Logger.Infof("Updated %d playlists", playlistCount)
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the daemon to scrape now playing songs periodically",
	Run:   runDaemon,
}

func init() {
	daemonCmd.Flags().BoolVar(&noStore, "no-store", false, "Run without storing the now playing songs")
	daemonCmd.Flags().BoolVar(&noPlaylist, "no-playlist", false, "Run without updating the Spotify playlist")
	daemonCmd.Flags().DurationVar(&fetchInterval, "fetch-interval", 1*time.Minute, "Interval between scrapes (e.g., 30s, 1m, 5m)")
	daemonCmd.Flags().DurationVar(&playlistUpdateInterval, "playlist-update-interval", 1*time.Hour, "Interval between playlist updates (e.g., 30m, 1h, 5h)")
	daemonCmd.Flags().StringVar(&playlistRange, "playlist-range", "lastday", "Time range for playlist update (lasthour, lastday, lastweek)")
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon(cmd *cobra.Command, args []string) {
	utils.Logger.Info("Starting daemon")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	configHandler, err := utils.NewConfigHandler(stationFile)
	if err != nil {
		utils.Logger.Fatalf("Error loading config: %v", err)
	}

	store, err := storage.NewStorage(storageType, storagePath)
	if err != nil {
		utils.Logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		utils.Logger.Fatalf("Error initializing storage: %v", err)
	}

	var spotifyService *spotify.SpotifyService
	if !noPlaylist {
		spotifyService, err = spotify.NewSpotifyService(configHandler, store)
		if err != nil {
			utils.Logger.Fatalf("Error initializing Spotify service: %v", err)
		}
	} else {
		utils.Logger.Info("Running without Spotify playlist update")
	}

	scraperService := &ScraperService{
		FetchInterval:            fetchInterval,            // Use the fetch interval from the flag
		PlaylistUpdateInterval:   playlistUpdateInterval,   // Use the playlist update interval from the flag
		SessionKeepAliveInterval: sessionKeepAliveInterval, // Use the session keep alive interval from the flag
		stopScraper:              make(chan struct{}),
		configHandler:            configHandler,
		storage:                  store,
		spotify:                  spotifyService,
	}

	if healthCheckEnv := utils.GetEnv("ENABLE_HEALTHCHECK", "false"); strings.ToLower(healthCheckEnv) == "true" {
		portEnv := utils.GetEnv("HEALTHCHECK_PORT", "8585")
		port, err := strconv.Atoi(portEnv)
		if err != nil {
			utils.Logger.Fatalf("Error parsing healthcheck port: %v", err)
		}
		if noPlaylist {
			playlistUpdateInterval = 0
		}
		go utils.StartHealthCheckServer(port, fetchInterval, playlistUpdateInterval, spotifyService)
	}
	go scraperService.Start()

	<-stop

	utils.Logger.Info("Received interrupt signal")

	scraperService.Stop()
	utils.Logger.Info("Stopped daemon")
}
