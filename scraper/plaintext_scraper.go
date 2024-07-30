package scraper

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"

	"github.com/sirupsen/logrus"
)

type PlaintextScraper struct {
	*BaseScraper
	Regex *regexp.Regexp
}

func NewPlaintextScraper(logger *logrus.Logger, URL string, regex string) (*PlaintextScraper, error) {
	compiledRegex, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	return &PlaintextScraper{
		BaseScraper: NewBaseScraper(logger, URL),
		Regex:       compiledRegex,
	}, nil
}

func (s *PlaintextScraper) GetNowPlaying() (*Song, error) {
	resp, err := http.Get(s.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching URL: %s, status code: %d", s.URL, resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		matches := s.Regex.FindStringSubmatch(line)
		if len(matches) == 3 {
			return &Song{
				Artist: matches[1],
				Title:  matches[2],
			}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("no matches found")
}
