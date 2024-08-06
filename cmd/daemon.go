package cmd

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"radio-to-spotify/config"
	"radio-to-spotify/scraper"
	"radio-to-spotify/spotify"
	"radio-to-spotify/storage"

	"github.com/sirupsen/logrus"
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
	configHandler            *config.ConfigHandler
	storage                  storage.Storage
	spotify                  *spotify.SpotifyService
	logger                   *logrus.Logger
}

func (s *ScraperService) Start() {
	s.logger.Infof("Starting scraper service with fetch interval %v", s.FetchInterval)
	fetchTicker := time.NewTicker(s.FetchInterval)
	var playlistUpdateTicker, sessionKeepAliveTicker *time.Ticker

	if !noPlaylist {
		s.logger.Debugf("Starting playlist update ticker with interval %v", s.PlaylistUpdateInterval)
		playlistUpdateTicker = time.NewTicker(s.PlaylistUpdateInterval)
		s.logger.Debugf("Starting session keep alive ticker with interval %v", s.SessionKeepAliveInterval)
		sessionKeepAliveTicker = time.NewTicker(s.SessionKeepAliveInterval)
	} else { // If no playlist update, don't start the tickers / Kinda hacky
		playlistUpdateTicker = time.NewTicker(1)
		sessionKeepAliveTicker = time.NewTicker(1)
		playlistUpdateTicker.Stop()
		sessionKeepAliveTicker.Stop()
	}

	for {
		select {
		case <-fetchTicker.C:
			s.fetchNowPlaying()
		case <-playlistUpdateTicker.C:
			s.updatePlaylists()
		case <-sessionKeepAliveTicker.C:
			s.keepSpotifySessionAlive()
		case <-s.stopScraper:
			fetchTicker.Stop()
			if playlistUpdateTicker != nil {
				playlistUpdateTicker.Stop()
			}
			if sessionKeepAliveTicker != nil {
				sessionKeepAliveTicker.Stop()
			}
			return
		}
	}
}

func (s *ScraperService) Stop() {
	s.logger.Info("Stopping scraper service")
	close(s.stopScraper)
}

func (s *ScraperService) fetchNowPlaying() {
	s.logger.Debugf("Fetching now playing songs")
	var storedCount, songCount int

	if !noStore {
		stations, songs, err := scraper.FetchNowPlaying(s.configHandler, s.logger, stationID)
		if err != nil {
			s.logger.Warnf("Error fetching now playing: %v", err)
			return
		}
		songCount = len(songs)

		for i, station := range stations {
			err := s.storage.StoreNowPlaying(station.ID, songs[i])
			if err != nil {
				s.logger.Errorf("Error storing now playing for station %s: %v", station.ID, err)
			} else {
				s.logger.Debugf("Stored song for station %s: %s - %s", station.ID, songs[i].Artist, songs[i].Title)
				storedCount++
			}
		}
	}

	s.logger.Infof("Fetched %d stations, stored %d songs", songCount, storedCount)
}

func (s *ScraperService) updatePlaylists() {
	if noPlaylist {
		return
	}

	s.logger.Debugf("Updating playlists")
	var playlistCount int
	var wg sync.WaitGroup

	stations, err := s.storage.GetAllStations()
	if err != nil {
		s.logger.Errorf("Error getting all stations: %v", err)
		return
	}
	if stationID != "" {
		stations = []string{stationID}
	} else {
		s.logger.Debugf("Updating playlists for all stations")
	}

	for _, stationID := range stations {
		wg.Add(1)
		go func(stationID string) {
			defer wg.Done()
			err := s.spotify.UpdateSpotifyPlaylist(stationID, playlistRange)
			if err != nil {
				s.logger.Errorf("Error updating Spotify playlist for station %s: %v", stationID, err)
			} else {
				playlistCount++
			}
		}(stationID)
	}
	wg.Wait()

	s.logger.Infof("Updated %d playlists", playlistCount)
}

func (s *ScraperService) keepSpotifySessionAlive() {
	s.logger.Debug("Keeping Spotify session alive")
	err := s.spotify.UpdateSession()
	if err != nil {
		s.logger.Errorf("Error keeping Spotify session alive: %v", err)
	} else {
		s.logger.Debug("Spotify session is active")
	}
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
	daemonCmd.Flags().DurationVar(&sessionKeepAliveInterval, "session-keep-alive-interval", 5*time.Minute, "Interval to keep the Spotify session alive")
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon(cmd *cobra.Command, args []string) {
	logger.Info("Starting daemon")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	configHandler, err := config.NewConfigHandler(stationFile)
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	store, err := storage.NewStorage(storageType, storagePath)
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	var spotifyService *spotify.SpotifyService
	if !noPlaylist {
		spotifyService, err = spotify.NewSpotifyService(logger, configHandler, store)
		if err != nil {
			logger.Fatalf("Error initializing Spotify service: %v", err)
		}
	} else {
		logger.Info("Running without Spotify playlist update")
	}

	scraperService := &ScraperService{
		FetchInterval:            fetchInterval,            // Use the fetch interval from the flag
		PlaylistUpdateInterval:   playlistUpdateInterval,   // Use the playlist update interval from the flag
		SessionKeepAliveInterval: sessionKeepAliveInterval, // Use the session keep alive interval from the flag
		stopScraper:              make(chan struct{}),
		configHandler:            configHandler,
		storage:                  store,
		spotify:                  spotifyService,
		logger:                   logger,
	}

	go scraperService.Start()

	<-stop

	logger.Info("Received interrupt signal")

	scraperService.Stop()
	logger.Info("Stopped daemon")
}
