package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"radio-to-spotify/storage"
)

var (
	logger      = logrus.New()
	storageType string
	storagePath string
	store       storage.Storage
	logLevel    string
	stationFile string
	stationID   string
)

var rootCmd = &cobra.Command{
	Use:   "radio-to-spotify",
	Short: "Radio to Spotify is a tool to fetch now playing songs from radio stations and create Spotify playlists.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&storageType, "storage", "file", "Storage type: file or sqlite")
	rootCmd.PersistentFlags().StringVar(&storagePath, "storage-path", "data", "Path to storage file or database")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Logging level: debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().StringVar(&stationFile, "station-file", "stations.json", "Path to stations file")
	rootCmd.PersistentFlags().StringVar(&stationID, "station", "", "Station ID to fetch/store now playing")

	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(playlistCmd)
}

func initConfig() {
	var err error
	store, err = storage.NewStorage(storageType, storagePath)
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	level, err := logrus.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		logger.Fatalf("Invalid log level: %v", err)
	}
	logger.SetLevel(level)
}
