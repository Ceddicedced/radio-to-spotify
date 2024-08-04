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
	interval                 time.Duration
	playlistRange            string
	sessionKeepAliveInterval time.Duration
)

type ScraperService struct {
	Interval                 time.Duration
	SessionKeepAliveInterval time.Duration
	stopScraper              chan struct{}
	configHandler            *config.ConfigHandler
	storage                  storage.Storage
	spotify                  *spotify.SpotifyService
	logger                   *logrus.Logger
}

func (s *ScraperService) Start() {
	s.logger.Infof("Starting scraper service with interval %v", s.Interval)
	ticker := time.NewTicker(s.Interval)
	var sessionKeepAliveTicker *time.Ticker
	if !noPlaylist {
		s.logger.Debugf("Starting session keep alive ticker with interval %v", s.SessionKeepAliveInterval)
		sessionKeepAliveTicker = time.NewTicker(s.SessionKeepAliveInterval)
	}

	for {
		select {
		case <-ticker.C:
			s.scrape()
		case <-sessionKeepAliveTicker.C:
			s.keepSpotifySessionAlive()
		case <-s.stopScraper:
			ticker.Stop()
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

func (s *ScraperService) scrape() {
	s.logger.Debugf("Scraping now playing songs")
	var storedCount, playlistCount, songCount int
	var wg sync.WaitGroup

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

	if !noPlaylist {
		stations, err := s.storage.GetAllStations()
		if err != nil {
			s.logger.Errorf("Error getting all stations: %v", err)
			return
		}
		if stationID != "" {
			stations = []string{stationID}
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
	}

	s.logger.Infof("Scraped %d stations, stored %d songs, updated %d playlists", songCount, storedCount, playlistCount)
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
	daemonCmd.Flags().DurationVar(&interval, "interval", 1*time.Minute, "Interval between scrapes (e.g., 30s, 1m, 5m)")
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
		Interval:                 interval,                 // Use the interval from the flag
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
