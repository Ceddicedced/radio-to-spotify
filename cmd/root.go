package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"radio-to-spotify/storage"
	"radio-to-spotify/utils"
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

	rootCmd.PersistentFlags().StringVar(&storageType, "storage", "sqlite", "Storage type: file or sqlite or postgres")
	rootCmd.PersistentFlags().StringVar(&storagePath, "storage-path", "data", "Path to storage file or database connection string")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Logging level: debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().StringVar(&stationFile, "station-file", "stations.json", "Path to stations file")
	rootCmd.PersistentFlags().StringVar(&stationID, "station", "", "Station ID to fetch/store now playing")
}

func initConfig() {
	var err error

	utils.SetLevel(logLevel)
	store, err = storage.NewStorage(storageType, storagePath)
	if err != nil {
		utils.Logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		utils.Logger.Fatalf("Error initializing storage: %v", err)
	}

}
