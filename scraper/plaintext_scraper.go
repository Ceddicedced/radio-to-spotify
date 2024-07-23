package scraper

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html/charset"
)

type PlaintextScraper struct {
	BaseScraper
	regex *regexp.Regexp
}

func NewPlaintextScraper(logger *logrus.Logger, url, regexPattern string) (*PlaintextScraper, error) {
	compiledRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex pattern: %v", err)
	}

	return &PlaintextScraper{
		BaseScraper: *NewBaseScraper(logger, url),
		regex:       compiledRegex,
	}, nil
}

func (p *PlaintextScraper) GetNowPlaying() (*Song, error) {
	p.Logger.Infof("Fetching plaintext now playing from URL: %s", p.BaseScraper.URL)
	res, err := http.Get(p.BaseScraper.URL)
	if err != nil {
		p.Logger.Errorf("Error fetching plaintext now playing: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		p.Logger.Errorf("Received non-200 status code: %d", res.StatusCode)
		return nil, err
	}

	reader, err := charset.NewReader(res.Body, res.Header.Get("Content-Type"))
	if err != nil {
		p.Logger.Errorf("Error creating reader for response body: %v", err)
		return nil, err
	}

	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		p.Logger.Errorf("Error reading plaintext response: %v", err)
		return nil, err
	}

	text := scanner.Text()
	matches := p.regex.FindStringSubmatch(text)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found in plaintext response")
	}

	result := make(map[string]string)
	for i, name := range p.regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = matches[i]
		}
	}

	artist, artistOk := result["artist"]
	title, titleOk := result["title"]
	if !artistOk || !titleOk {
		return nil, fmt.Errorf("artist or title not found in plaintext response")
	}

	return &Song{Artist: artist, Title: title}, nil
}
