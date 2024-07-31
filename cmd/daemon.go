package cmd

import (
	"os"
	"os/signal"
	"time"

	"radio-to-spotify/config"
	"radio-to-spotify/scraper"
	"radio-to-spotify/spotify"
	"radio-to-spotify/storage"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	noStore    bool
	noPlaylist bool
	interval   time.Duration
)

type ScraperService struct {
	Interval      time.Duration
	stopScraper   chan struct{}
	configHandler *config.ConfigHandler
	storage       storage.Storage
	spotify       *spotify.SpotifyService
	logger        *logrus.Logger
}

func (s *ScraperService) Start() {
	s.logger.Infof("Starting scraper service with interval %v", s.Interval)
	ticker := time.NewTicker(s.Interval)
	for {
		select {
		case <-ticker.C:
			s.scrape()
		case <-s.stopScraper:
			ticker.Stop()
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
	stations, songs, err := scraper.FetchNowPlaying(s.configHandler, s.logger, stationID)
	if err != nil {
		s.logger.Warnf("Error fetching now playing: %v", err)
		return
	}

	var storedCount, playlistCount int
	for i, station := range stations {
		if !noStore {
			err := s.storage.StoreNowPlaying(station.ID, songs[i])
			if err != nil {
				s.logger.Errorf("Error storing now playing for station %s: %v", station.ID, err)
			} else {
				s.logger.Debugf("Stored song for station %s: %s - %s", station.ID, songs[i].Artist, songs[i].Title)
				storedCount++
			}
		}

		if !noPlaylist {
			err = s.spotify.UpdateSpotifyPlaylist(station.ID)
			if err != nil {
				s.logger.Errorf("Error updating Spotify playlist for station %s: %v", station.ID, err)
			} else {
				s.logger.Debugf("Updated Spotify playlist for station: %s", station.Name)
				playlistCount++
			}
		}

	}

	s.logger.Infof("Scraped %d stations, stored %d songs, updated %d playlists", len(stations), storedCount, playlistCount)
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
		Interval:      interval, // Use the interval from the flag
		stopScraper:   make(chan struct{}),
		configHandler: configHandler,
		storage:       store,
		spotify:       spotifyService,
		logger:        logger,
	}

	go scraperService.Start()

	<-stop

	logger.Info("Received interrupt signal")

	scraperService.Stop()
	logger.Info("Stopped daemon")
}
