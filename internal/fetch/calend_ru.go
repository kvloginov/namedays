package fetch

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kvloginov/namedays/internal/domain"
)

// CalendFetcher fetches namedays for a specific date
type CalendFetcher struct {
	baseURL string
	client  *http.Client
}

// NewCalendFetcher creates a new instance of NamedaysFetcher
func NewCalendFetcher() *CalendFetcher {
	return &CalendFetcher{
		baseURL: "https://www.calend.ru/names",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *CalendFetcher) FetchAllNamedays() (domain.NamedaysDataList, error) {
	now := time.Now()
	currentYear := now.Year()

	startDate := time.Date(currentYear, time.January, 1, 0, 0, 0, 0, time.Local)

	namedays := domain.NamedaysDataList{}

	for day := 0; day < 365; day++ {
		date := startDate.AddDate(0, 0, day)
		names, err := f.fetchNamedays(date)
		if err != nil {
			return nil, fmt.Errorf("error fetching namedays: %w", err)
		}

		namedays = append(namedays, domain.NamedaysData{
			Date:  domain.NewDayMonth(date),
			Names: names,
		})
	}

	return namedays, nil
}

// FetchNamedays fetches namedays for a specific date
func (f *CalendFetcher) fetchNamedays(date time.Time) ([]string, error) {
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
