package scraper

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"
)

type PlaintextScraper struct {
	*BaseScraper
	Regex *regexp.Regexp
}

func NewPlaintextScraper(URL string, regex string) (*PlaintextScraper, error) {
	compiledRegex, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	return &PlaintextScraper{
		BaseScraper: NewBaseScraper(URL),
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
			// Clean up artist and title by filtering out invalid values like NULL or \x00
			artist := cleanString(matches[1])
			title := cleanString(matches[2])

			// Skip if artist or title is empty or contains invalid characters
			if artist == "" || title == "" {
				continue
			}

			return &Song{
				Artist: artist,
				Title:  title,
			}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("no matches found")
}

// cleanString filters out invalid or unwanted characters like NULL, \x00, etc.
func cleanString(s string) string {
	// Define a regex pattern to remove invalid characters like \x00 or NULL
	re := regexp.MustCompile(`(?i)(\x00|null)`)
	return re.ReplaceAllString(s, "")
}
