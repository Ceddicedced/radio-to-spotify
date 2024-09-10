package scraper

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type HTMLScraper struct {
	*BaseScraper
	ArtistTag string
	TitleTag  string
}

func NewHTMLScraper(URL string, artistTag string, titleTag string) *HTMLScraper {
	return &HTMLScraper{
		BaseScraper: NewBaseScraper(URL),
		ArtistTag:   artistTag,
		TitleTag:    titleTag,
	}
}

func (s *HTMLScraper) GetNowPlaying() (*Song, error) {
	resp, err := http.Get(s.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching URL: %s, status code: %d", s.URL, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	artist := doc.Find(s.ArtistTag).Text()
	title := doc.Find(s.TitleTag).Text()

	if artist == "" || title == "" {
		return nil, fmt.Errorf("could not find artist or title in HTML")
	}

	return &Song{
		Artist: artist,
		Title:  title,
	}, nil
}
