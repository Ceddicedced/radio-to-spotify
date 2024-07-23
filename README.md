[![Go](https://github.com/Ceddicedced/radio-to-spotify/actions/workflows/go.yml/badge.svg)](https://github.com/Ceddicedced/radio-to-spotify/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
# Radio to Spotify 🎵

`radio-to-spotify` is a Go-based tool that scrapes now-playing songs from online radio stations and creates Spotify playlists based on the last hour of songs played by a station. It can run as a daemon to periodically fetch and store the now-playing data.

## Features
- Fetch now-playing songs from various radio stations
- Store fetched songs in a file or SQLite database
- Create Spotify playlists based on the stored songs
- Run as a daemon to periodically fetch and store songs

## Installation

### Prerequisites
- [Go](https://golang.org/doc/install) (version 1.16+)
- [SQLite](https://www.sqlite.org/index.html) (if using SQLite storage)
- [Spotify Developer Account](https://developer.spotify.com/)

### Clone the Repository
```sh
git clone https://github.com/yourusername/radio-to-spotify.git
cd radio-to-spotify
```

### Build the Project
```sh
go build -o radio-to-spotify main.go
```

## Configuration

Create a `stations.json` file to define the radio stations to scrape:
```json
{
  "stations": [
    {
      "id": "radiofritz",
      "name": "Radio Fritz",
      "url": "https://www.fritz.de/include/frz/nowonair/now_on_air.html",
      "type": "html",
      "artistTag": "p.artist",
      "titleTag": "p.songtitle"
    },
    {
      "id": "njoy",
      "name": "N-JOY",
      "url": "https://www.n-joy.de/public/radioplaylists/njoy.json",
      "type": "json",
      "artistKey": ["song_now_interpret"],
      "titleKey": ["song_now_title"]
    }
  ]
}
```

## Usage

### Fetch Now Playing
Fetch the now-playing songs for all stations defined in the `config.json` file:
```sh
./radio-to-spotify fetch --config=config.json --loglevel=debug --storage=file --storage-path=data/db.json
```

### Store Now Playing
Fetch and store now-playing songs for a specific station (dry run):
```sh
./radio-to-spotify store --config=config.json --station=radiofritz --loglevel=info --dry-run --storage=file --storage-path=data/db.json
```

### Create Spotify Playlist
Create a Spotify playlist for the last hour of songs for a specific station:
```sh
./radio-to-spotify playlist --config=config.json --station=radiofritz --loglevel=error --storage=file --storage-path=data/db.json
```

### Run as a Daemon
Run the tool as a daemon to periodically fetch and store now-playing songs:
```sh
./radio-to-spotify daemon --config=config.json --loglevel=debug --storage=file --storage-path=data/db.json
```

## Environment Variables
For Spotify integration, set the following environment variables:
- `SPOTIFY_ID`: Your Spotify Client ID
- `SPOTIFY_SECRET`: Your Spotify Client Secret
- `SPOTIFY_REDIRECT_URL`: Your Spotify Redirect URL

```sh
export SPOTIFY_ID=your_spotify_client_id
export SPOTIFY_SECRET=your_spotify_client_secret
export SPOTIFY_REDIRECT_URL=your_spotify_redirect_url
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any changes.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgements
- [Go](https://golang.org/)
- [Cobra](https://github.com/spf13/cobra)
- [Logrus](https://github.com/sirupsen/logrus)
- [Spotify Web API](https://developer.spotify.com/documentation/web-api/)

## Contact
Created by [Ceddicedced](https://github.com/ceddicedced) - feel free to contact me!

