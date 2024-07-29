package scraper

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

type HTMLScraper struct {
	BaseScraper
	artistTag string
	titleTag  string
}

func NewHTMLScraper(logger *logrus.Logger, url, artistTag, titleTag string) *HTMLScraper {
	return &HTMLScraper{
		BaseScraper: *NewBaseScraper(logger, url),
		artistTag:   artistTag,
		titleTag:    titleTag,
	}
}

func (h *HTMLScraper) GetNowPlaying() (*Song, error) {
	h.Logger.Debugf("Fetching HTML now playing from URL: %s", h.BaseScraper.URL)
	res, err := http.Get(h.BaseScraper.URL)
	if err != nil {
		h.Logger.Errorf("Error fetching HTML now playing: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		h.Logger.Errorf("Received non-200 status code: %d", res.StatusCode)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		h.Logger.Errorf("Error parsing HTML document: %v", err)
		return nil, err
	}

	artist := doc.Find(h.artistTag).First().Text()
	title := doc.Find(h.titleTag).First().Text()

	if artist == "" || title == "" {
		return nil, fmt.Errorf("could not find now playing song details")
	}

	return &Song{Artist: artist, Title: title}, nil
}
