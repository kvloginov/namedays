package parser

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// NamedaysFetcher fetches namedays for a specific date
type NamedaysFetcher struct {
	baseURL string
	client  *http.Client
}

// NewNamedaysFetcher creates a new instance of NamedaysFetcher
func NewNamedaysFetcher() *NamedaysFetcher {
	return &NamedaysFetcher{
		baseURL: "https://www.calend.ru/names",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchNamedays fetches namedays for a specific date
func (f *NamedaysFetcher) FetchNamedays(date time.Time) ([]string, error) {
	url := fmt.Sprintf("%s/%d-%d-%d/", f.baseURL, date.Year(), date.Month(), date.Day())

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching namedays: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	var names []string
	doc.Find("a.title.name.M, a.title.name.F").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		if name != "" {
			names = append(names, name)
		}
	})

	return names, nil
}

// FormatDateKey formats date as MMDD string
func FormatDateKey(date time.Time) string {
	return fmt.Sprintf("%02d%02d", date.Month(), date.Day())
}
