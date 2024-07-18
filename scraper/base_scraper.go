package scraper

import (
	"github.com/sirupsen/logrus"
)

type Song struct {
	Time   string
	Artist string
	Title  string
}

type Station struct {
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Type      string        `json:"type"`
	ArtistTag string        `json:"artistTag,omitempty"`
	TitleTag  string        `json:"titleTag,omitempty"`
	ArtistKey []interface{} `json:"artistKey,omitempty"`
	TitleKey  []interface{} `json:"titleKey,omitempty"`
}

type Scraper interface {
	GetNowPlaying() (*Song, error)
}

type BaseScraper struct {
	Logger *logrus.Logger
	URL    string
}

func NewBaseScraper(logger *logrus.Logger, URL string) *BaseScraper {
	return &BaseScraper{Logger: logger, URL: URL}
}
