/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"radio-to-spotify/cmd"

	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	cmd.Execute()
}

// Load environment variables from .env file if it exists
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}
