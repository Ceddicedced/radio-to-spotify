package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONScraper struct {
	*BaseScraper
	ArtistKey []interface{}
	TitleKey  []interface{}
}

func NewJSONScraper(URL string, artistKey []interface{}, titleKey []interface{}) *JSONScraper {
	return &JSONScraper{
		BaseScraper: NewBaseScraper( URL),
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

	var data interface{}
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

	artistStr, ok := artist.(string)
	if !ok {
		return nil, fmt.Errorf("artist key does not contain a string")
	}

	titleStr, ok := title.(string)
	if !ok {
		return nil, fmt.Errorf("title key does not contain a string")
	}

	return &Song{
		Artist: artistStr,
		Title:  titleStr,
	}, nil
}

func getValueFromKey(data interface{}, keys []interface{}) (interface{}, error) {
	var value interface{} = data
	for _, key := range keys {
		switch key := key.(type) {
		case string:
			if m, ok := value.(map[string]interface{}); ok {
				value = m[key]
			} else {
				return nil, fmt.Errorf("invalid key type: %T", key)
			}
		case int:
			if a, ok := value.([]interface{}); ok {
				value = a[key]
			} else {
				return nil, fmt.Errorf("invalid key type: %T", key)
			}
		case float64:
			index := int(key)
			if a, ok := value.([]interface{}); ok {
				value = a[index]
			} else {
				return nil, fmt.Errorf("invalid key type: %T", key)
			}
		default:
			return nil, fmt.Errorf("invalid key type: %T", key)
		}
	}
	return value, nil
}
