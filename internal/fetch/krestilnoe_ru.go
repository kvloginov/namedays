package fetch

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kvloginov/namedays/internal/domain"
)

type KrestilnoeFetcher struct {
	baseURL string
	client  *http.Client
}

func NewKrestilnoeFetcher() *KrestilnoeFetcher {
	return &KrestilnoeFetcher{
		baseURL: "https://www.krestilnoe.ru/svyattsy-kalendar-god/",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *KrestilnoeFetcher) FetchAllNamedays() (domain.NamedaysDataList, error) {
	resp, err := f.client.Get(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching krestilnoe.ru: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	namedays := domain.NamedaysDataList{}
	currentYear := time.Now().Year()

	// Find all paragraphs with calendar data
	// Data is inside <p> tags with formatting through <br>
	doc.Find("p").Each(func(i int, paragraph *goquery.Selection) {
		// Get HTML content of the paragraph
		html, _ := paragraph.Html()
		// Check if the paragraph contains calendar data (first two characters - number and space)
		if len(html) > 2 && regexp.MustCompile(`^\d+\s`).MatchString(html) {
			// Find the first month in the text to determine which month this block belongs to
			monthMatch := regexp.MustCompile(`\d+\s+([^\s:]+):`).FindStringSubmatch(html)
			if len(monthMatch) > 1 {
				monthName := monthMatch[1]
				monthNum := extractMonthNumber(monthName)

				if monthNum > 0 {
					// Parse dates and names for this month
					monthNamedays := parseMonthNamedays(html, monthNum, currentYear)
					namedays = append(namedays, monthNamedays...)
				}
			}
		}
	})

	return namedays, nil
}

// extractMonthNumber gets the month number from the month title
func extractMonthNumber(monthTitle string) int {
	monthMap := map[string]int{
		"января":   1,
		"январе":   1,
		"февраля":  2,
		"феврале":  2,
		"марта":    3,
		"марте":    3,
		"апреля":   4,
		"апреле":   4,
		"мая":      5,
		"мае":      5,
		"июня":     6,
		"июне":     6,
		"июля":     7,
		"июле":     7,
		"августа":  8,
		"августе":  8,
		"сентября": 9,
		"сентябре": 9,
		"октября":  10,
		"октябре":  10,
		"ноября":   11,
		"ноябре":   11,
		"декабря":  12,
		"декабре":  12,
	}

	for month, num := range monthMap {
		if strings.Contains(strings.ToLower(monthTitle), month) {
			return num
		}
	}

	return 0
}

// parseMonthNamedays parses the text with names for a month
func parseMonthNamedays(text string, monthNum, year int) []domain.NamedaysData {
	// HTML contains <br> between days, which we need to correctly process
	// First, clean the text from extra formatting and normalize spaces
	text = strings.TrimSpace(text)

	// Split the text into days by <br> tags
	dayEntries := strings.Split(text, "<br/>")

	result := []domain.NamedaysData{}

	for _, entry := range dayEntries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Regular expression to extract the day and names
		// Format: "1 января: Илья, Вонифатий, ..."
		re := regexp.MustCompile(`^(\d+)\s+[^:]+:\s+(.+)$`)

		matches := re.FindStringSubmatch(entry)
		if len(matches) < 3 {
			continue
		}

		day, err := strconv.Atoi(matches[1])
		if err != nil || day < 1 || day > 31 {
			continue
		}

		// Create the date
		date := time.Date(year, time.Month(monthNum), day, 0, 0, 0, 0, time.Local)

		// Extract names, separated by commas
		namesStr := matches[2]
		namesArr := strings.Split(namesStr, ",")

		// Clean and filter names
		var names []string
		for _, name := range namesArr {
			name = strings.TrimSpace(name)
			// Remove "and others" and empty values
			if name != "" && name != "и иные" && name != "и др." &&
				!strings.Contains(name, " января:") &&
				!strings.Contains(name, " февраля:") &&
				!strings.Contains(name, " марта:") &&
				!strings.Contains(name, " апреля:") &&
				!strings.Contains(name, " мая:") &&
				!strings.Contains(name, " июня:") &&
				!strings.Contains(name, " июля:") &&
				!strings.Contains(name, " августа:") &&
				!strings.Contains(name, " сентября:") &&
				!strings.Contains(name, " октября:") &&
				!strings.Contains(name, " ноября:") &&
				!strings.Contains(name, " декабря:") {
				names = append(names, name)
			}
		}

		if len(names) > 0 {
			result = append(result, domain.NamedaysData{
				Date:  domain.NewDayMonth(date),
				Names: names,
			})
		}
	}

	return result
}
