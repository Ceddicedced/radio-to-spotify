package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/sirupsen/logrus"
)

type JSONScraper struct {
	BaseScraper
	artistKey []interface{}
	titleKey  []interface{}
}

func NewJSONScraper(logger *logrus.Logger, url string, artistKey, titleKey []interface{}) *JSONScraper {
	return &JSONScraper{
		BaseScraper: *NewBaseScraper(logger, url),
		artistKey:   artistKey,
		titleKey:    titleKey,
	}
}

func getValueFromPath(data interface{}, path []interface{}) (string, error) {
	var current interface{} = data
	for _, key := range path {
		switch k := key.(type) {
		case string:
			if currentMap, ok := current.(map[string]interface{}); ok {
				current = currentMap[k]
			} else {
				return "", fmt.Errorf("expected map at %v but got %T", k, current)
			}
		case float64: // JSON numbers are unmarshalled as float64 by default /tableflip
			index := int(k)
			if currentArray, ok := current.([]interface{}); ok {
				if index < len(currentArray) {
					current = currentArray[index]
				} else {
					return "", fmt.Errorf("index %d out of bounds", index)
				}
			} else {
				return "", fmt.Errorf("expected array at %v but got %T", k, current)
			}
		default:
			return "", fmt.Errorf("unexpected key type %T", key)
		}
	}
	if value, ok := current.(string); ok {
		return value, nil
	}
	return "", fmt.Errorf("value at path is not a string, got %s", reflect.TypeOf(current).String())
}

func (j *JSONScraper) GetNowPlaying() (*Song, error) {
	j.Logger.Infof("Fetching JSON now playing from URL: %s", j.BaseScraper.URL)
	res, err := http.Get(j.BaseScraper.URL)
	if err != nil {
		j.Logger.Errorf("Error fetching JSON now playing: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		j.Logger.Errorf("Received non-200 status code: %d", res.StatusCode)
		return nil, err
	}

	var jsonData interface{}
	err = json.NewDecoder(res.Body).Decode(&jsonData)
	if err != nil {
		j.Logger.Errorf("Error parsing JSON document: %v", err)
		return nil, err
	}

	artist, errArtist := getValueFromPath(jsonData, j.artistKey)
	title, errTitle := getValueFromPath(jsonData, j.titleKey)

	if errArtist != nil || errTitle != nil {
		return nil, fmt.Errorf("error fetching now playing song: artist error: %v, title error: %v", errArtist, errTitle)
	}

	return &Song{Artist: artist, Title: title}, nil
}
