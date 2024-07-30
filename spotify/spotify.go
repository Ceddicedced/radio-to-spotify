package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"radio-to-spotify/config"
	"radio-to-spotify/storage"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	clientID      = os.Getenv("SPOTIFY_ID")
	clientSecret  = os.Getenv("SPOTIFY_SECRET")
	redirectURL   = os.Getenv("SPOTIFY_REDIRECT_URL")
	port          = os.Getenv("SPOTIFY_PORT")
	authenticator = spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithRedirectURL(redirectURL),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
			spotifyauth.ScopePlaylistModifyPrivate,
		),
	)
	tokenFile = ".token"
)

func getAuthToken() (*oauth2.Token, error) {
	token, err := loadTokenFromFile(tokenFile)
	if err == nil && token.Valid() {
		return token, nil
	}

	http.HandleFunc("/callback", completeAuth)
	go http.ListenAndServe(":"+port, nil)

	url := authenticator.AuthURL("state-token")
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	for {
		time.Sleep(100 * time.Millisecond)
		token, err := loadTokenFromFile(tokenFile)
		if err == nil && token.Valid() {
			return token, nil
		}
	}
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := authenticator.Token(context.Background(), "state-token", r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		return
	}
	if st := r.FormValue("state"); st != "state-token" {
		http.NotFound(w, r)
		return
	}

	saveTokenToFile(tokenFile, tok)

	client := spotify.New(authenticator.Client(context.Background(), tok))
	_, err = client.CurrentUser(context.Background())
	if err != nil {
		http.Error(w, "Couldn't get user", http.StatusForbidden)
		return
	}

	fmt.Fprintf(w, "Login Completed! You can close this window.")
}

func saveTokenToFile(path string, token *oauth2.Token) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(token)
}

func loadTokenFromFile(path string) (*oauth2.Token, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token oauth2.Token
	err = json.NewDecoder(file).Decode(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func getClient() (*spotify.Client, error) {
	token, err := getAuthToken()
	if err != nil {
		return nil, err
	}

	client := spotify.New(authenticator.Client(context.Background(), token))
	return client, nil
}

func CreateSpotifyPlaylist(configHandler *config.ConfigHandler, stationID string, store storage.Storage) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		return err
	}

	fmt.Printf("Logged in as: %s\n", user.DisplayName)

	station, err := configHandler.GetStationByID(stationID)
	if err != nil {
		return err
	}

	songs, err := store.GetLastHourSongs(stationID)
	if err != nil {
		return err
	}

	var playlistID spotify.ID
	if station.PlaylistID != "" {
		playlistID = spotify.ID(station.PlaylistID)
	} else {
		playlistName := fmt.Sprintf("Now Playing on %s - %s", station.Name, time.Now().Format("2006-01-02 15:04:05"))
		playlist, err := client.CreatePlaylistForUser(context.Background(), user.ID, playlistName, "Songs played on "+station.Name, false, false)
		if err != nil {
			return err
		}
		playlistID = playlist.ID
		station.PlaylistID = string(playlistID)
		err = configHandler.UpdateStation(station)
		if err != nil {
			return err
		}
	}

	var trackIDs []spotify.ID
	for _, song := range songs {
		searchResults, err := client.Search(context.Background(), fmt.Sprintf("%s %s", song.Artist, song.Title), spotify.SearchTypeTrack)
		if err != nil {
			return err
		}
		if searchResults.Tracks.Total > 0 {
			trackIDs = append(trackIDs, searchResults.Tracks.Tracks[0].ID)
		}
	}

	_, err = client.AddTracksToPlaylist(context.Background(), playlistID, trackIDs...)
	if err != nil {
		return err
	}

	fmt.Printf("Playlist updated: %s\n", station.Name)
	return nil
}
