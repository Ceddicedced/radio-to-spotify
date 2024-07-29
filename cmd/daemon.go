package cmd

import (
	"os"
	"os/signal"
	"time"

	"radio-to-spotify/scraper"
	"radio-to-spotify/storage"

	"github.com/spf13/cobra"
)

type ScraperService struct {
	Interval    time.Duration
	stopScraper chan struct{}
}

func (s *ScraperService) Start() {
	logger.Infof("Starting scraper service with interval %v", s.Interval)
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
	logger.Info("Stopping scraper service")
	close(s.stopScraper)
}

func (s *ScraperService) scrape() {
	logger.Debugf("Scraping now playing songs")
	stations, err := scraper.FetchNowPlaying(stationFile, logger, "")
	if err != nil {
		logger.Warnf("Error fetching now playing: %v", err)
		return
	}

	for _, stationSong := range stations {
		err := store.StoreNowPlaying(stationSong.StationID, &stationSong.Song)
		if err != nil {
			logger.Errorf("Error storing now playing for station %s: %v", stationSong.StationID, err)
		} else {
			logger.Debugf("Stored song for station %s: %s - %s", stationSong.StationID, stationSong.Song.Artist, stationSong.Song.Title)
		}
	}
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the daemon to scrape now playing songs periodically",
	Run:   runDaemon,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon(cmd *cobra.Command, args []string) {
	logger.Info("Starting daemon")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	store, err := storage.NewStorage(storageType, storagePath)
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	scraperService := &ScraperService{
		Interval:    1 * time.Minute, // Adjust the interval as needed
		stopScraper: make(chan struct{}),
	}

	go scraperService.Start()

	<-stop

	logger.Info("Received interrupt signal")

	scraperService.Stop()
	logger.Info("Stopped daemon")
}
