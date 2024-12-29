package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"radio-to-spotify/storage"
)

var (
	logger      = logrus.New()
	store       storage.Storage
	storageType string
	storagePath string
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

	// Define flags
	rootCmd.PersistentFlags().StringVar(&storageType, "storage", "", "Storage type: file, sqlite, or postgres")
	rootCmd.PersistentFlags().StringVar(&storagePath, "storage-path", "", "Path to storage file or database connection string")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "", "Logging level: debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().StringVar(&stationFile, "station-file", "", "Path to stations file")
	rootCmd.PersistentFlags().StringVar(&stationID, "station", "", "Station ID to fetch/store now playing")

	// Bind flags to Viper
	viper.BindPFlag("storage", rootCmd.PersistentFlags().Lookup("storage"))
	viper.BindPFlag("storage_path", rootCmd.PersistentFlags().Lookup("storage-path"))
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("station_file", rootCmd.PersistentFlags().Lookup("station-file"))
	viper.BindPFlag("station_id", rootCmd.PersistentFlags().Lookup("station"))

	// Set up Viper to read from environment variables
	viper.SetEnvPrefix("RTS") // Will look for environment variables with the prefix RTS_
	viper.AutomaticEnv()
}

func initConfig() {
	// Set default values in case none are provided via environment or flags
	viper.SetDefault("storage", "sqlite")
	viper.SetDefault("storage_path", "data")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("station_file", "stations.json")
	viper.SetDefault("station_id", "")

	// Read configuration from file if available (optional)
	viper.SetConfigName("config") // Name of config file (without extension)
	viper.SetConfigType("yaml")   // Specify the config file format (could be json, toml, etc.)
	viper.AddConfigPath(".")      // Look for the config file in the current directory

	// Attempt to read in the config file
	if err := viper.ReadInConfig(); err != nil {
		logger.Warnf("No config file found, using environment variables and defaults: %v", err)
	}

	// Initialize storage using Viper configuration
	storageType = viper.GetString("storage")
	storagePath = viper.GetString("storage_path")
	store, err := storage.NewStorage(storageType, storagePath)
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	err = store.Init()
	if err != nil {
		logger.Fatalf("Error initializing storage: %v", err)
	}

	logger.Infof("Using storage type: %s, path: %s", storageType, storagePath)
}
