package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type JSONScraper struct {
	*BaseScraper
	ArtistKey []interface{}
	TitleKey  []interface{}
}

func NewJSONScraper(logger *logrus.Logger, URL string, artistKey []interface{}, titleKey []interface{}) *JSONScraper {
	return &JSONScraper{
		BaseScraper: NewBaseScraper(logger, URL),
		ArtistKey:   artistKey,
		TitleKey:    titleKey,
	}
}

func (s *JSONScraper) GetNowPlaying() (*Song, error) {
	resp, err := http.Get(s.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching URL: %s, status code: %d", s.URL, resp.StatusCode)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	artist, err := getValueFromKey(data, s.ArtistKey)
	if err != nil {
		return nil, err
	}

	title, err := getValueFromKey(data, s.TitleKey)
	if err != nil {
		return nil, err
	}

	return &Song{
		Artist: artist.(string),
		Title:  title.(string),
	}, nil
}

func getValueFromKey(data map[string]interface{}, keys []interface{}) (interface{}, error) {
	var value interface{} = data
	for _, key := range keys {
		switch key := key.(type) {
		case string:
			value = value.(map[string]interface{})[key]
		case int:
			value = value.([]interface{})[key]
		default:
			return nil, fmt.Errorf("invalid key type: %T", key)
		}
	}
	return value, nil
}
