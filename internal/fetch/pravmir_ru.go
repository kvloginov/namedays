package fetch

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kvloginov/namedays/internal/domain"
)

// PravmirFetcher structure for parsing data from pravmir.ru
type PravmirFetcher struct {
	baseURL string
	client  *http.Client
}

// NewPravmirFetcher creates a new instance of PravmirFetcher
func NewPravmirFetcher() *PravmirFetcher {
	return &PravmirFetcher{
		baseURL: "https://www.pravmir.ru/pravoslavnyj-kalendar-imenin/",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchAllNamedays gets all namedays from pravmir.ru
func (f *PravmirFetcher) FetchAllNamedays() (domain.NamedaysDataList, error) {
	resp, err := f.client.Get(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching pravmir.ru: %w", err)
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

	// 1. Try to find data in tables
	tableNamedays := f.parseFromTables(doc, currentYear)
	namedays = append(namedays, tableNamedays...)

	// 2. If there's not enough data in tables, look in other formats
	if len(namedays) < 200 { // Expect more records for a full year
		// Search in text blocks of main content
		textBlockNamedays := f.parseFromTextBlocks(doc, currentYear)
		namedays = append(namedays, textBlockNamedays...)

		// Search in month blocks (in different possible formats)
		monthBlockNamedays := f.parseFromMonthBlocks(doc, currentYear)
		namedays = append(namedays, monthBlockNamedays...)
	}

	return namedays, nil
}

// parseFromTables tries to extract data from HTML tables
func (f *PravmirFetcher) parseFromTables(doc *goquery.Document, year int) domain.NamedaysDataList {
	result := domain.NamedaysDataList{}

	// Find and process tables with namedays
	doc.Find("table").Each(func(i int, tableElem *goquery.Selection) {
		// Check if this is a table with namedays
		if tableElem.Find("tr").Length() > 0 {
			tableElem.Find("tr").Each(func(j int, rowElem *goquery.Selection) {
				// Skip table headers
				if j == 0 && rowElem.Find("th").Length() > 0 {
					return
				}

				// First cell should contain the date
				dayCell := rowElem.Find("td").First()
				if dayCell.Length() == 0 {
					return
				}

				dayText := strings.TrimSpace(dayCell.Text())
				// Check if the first cell contains the date in "day month" format
				dateRe := regexp.MustCompile(`(\d+)\s*([а-яА-Я]+)`)
				matches := dateRe.FindStringSubmatch(dayText)

				if len(matches) < 3 {
					return
				}

				day := extractDay(matches[1])
				month := getMonthNumber(matches[2])

				if day == 0 || month == 0 {
					return
				}

				// Second cell contains names
				namesCell := rowElem.Find("td").Eq(1)
				if namesCell.Length() == 0 {
					return
				}

				namesText := strings.TrimSpace(namesCell.Text())
				names := parseNames(namesText)

				if len(names) > 0 {
					date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
					result = append(result, domain.NamedaysData{
						Date:  domain.NewDayMonth(date),
						Names: names,
					})
				}
			})
		}
	})

	return result
}

// parseFromTextBlocks tries to extract data from text blocks
func (f *PravmirFetcher) parseFromTextBlocks(doc *goquery.Document, year int) domain.NamedaysDataList {
	result := domain.NamedaysDataList{}

	// Search for blocks with namedays in the main content
	contentSelectors := []string{
		".entry-content p",
		".post-content p",
		".article-content p",
		".content p",
		"article p",
		".text p",
	}

	for _, selector := range contentSelectors {
		doc.Find(selector).Each(func(i int, p *goquery.Selection) {
			text := p.Text()

			// Format "day month: names"
			dateNameRe := regexp.MustCompile(`(\d+)\s+([а-яА-Я]+)[:]\s+(.+)`)
			matches := dateNameRe.FindAllStringSubmatch(text, -1)

			for _, match := range matches {
				if len(match) < 4 {
					continue
				}

				dayStr, monthStr, namesStr := match[1], match[2], match[3]

				day := extractDay(dayStr)
				month := getMonthNumber(monthStr)

				if day == 0 || month == 0 {
					continue
				}

				names := parseNames(namesStr)

				if len(names) > 0 {
					date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
					result = append(result, domain.NamedaysData{
						Date:  domain.NewDayMonth(date),
						Names: names,
					})
				}
			}
		})

		// If we found enough data, stop searching
		if len(result) > 50 {
			break
		}
	}

	return result
}

// parseFromMonthBlocks tries to extract data from month blocks
func (f *PravmirFetcher) parseFromMonthBlocks(doc *goquery.Document, year int) domain.NamedaysDataList {
	result := domain.NamedaysDataList{}

	// Search for month blocks (possible formats of the site)
	monthSelectors := []string{
		".month-block",
		".calendar-month",
		".month",
		"h2 + div", // Month header and next div
		"h3 + div",
		".namesBlock",
	}

	for _, selector := range monthSelectors {
		doc.Find(selector).Each(func(i int, monthBlock *goquery.Selection) {
			// Determine the month from the header or attribute
			var month int

			// Try to find the month in the text of the header
			monthTitle := monthBlock.Find("h2, h3, h4, .title").First().Text()
			if monthTitle != "" {
				month = getMonthNumber(monthTitle)
			}

			// If month is not found, try to find it in the text of the block itself
			if month == 0 {
				blockText := monthBlock.Text()
				for monthName, monthNum := range getMonthMap() {
					if strings.Contains(strings.ToLower(blockText), monthName) {
						month = monthNum
						break
					}
				}
			}

			// If month is not defined, skip the block
			if month == 0 {
				return
			}

			// Search for records about days and names in the block text
			content := monthBlock.Text()

			// Format: day with names (e.g. "1 января: Илья, Вонифатий, ...")
			dayNamesRe := regexp.MustCompile(`(\d+)[^:]*:\s*([^\.]+)`)
			dayMatches := dayNamesRe.FindAllStringSubmatch(content, -1)

			for _, match := range dayMatches {
				if len(match) < 3 {
					continue
				}

				dayStr, namesStr := match[1], match[2]

				day := extractDay(dayStr)
				if day == 0 {
					continue
				}

				names := parseNames(namesStr)

				if len(names) > 0 {
					date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
					result = append(result, domain.NamedaysData{
						Date:  domain.NewDayMonth(date),
						Names: names,
					})
				}
			}
		})

		// If we found enough data, stop searching
		if len(result) > 100 {
			break
		}
	}

	return result
}

// extractDay extracts the day number from a string
func extractDay(dayStr string) int {
	day := 0
	fmt.Sscanf(dayStr, "%d", &day)
	if day < 1 || day > 31 {
		return 0
	}
	return day
}

// getMonthMap returns a map of month names to their numbers
func getMonthMap() map[string]int {
	return map[string]int{
		"января":   1,
		"январь":   1,
		"январе":   1,
		"февраля":  2,
		"февраль":  2,
		"феврале":  2,
		"марта":    3,
		"март":     3,
		"марте":    3,
		"апреля":   4,
		"апрель":   4,
		"апреле":   4,
		"мая":      5,
		"май":      5,
		"мае":      5,
		"июня":     6,
		"июнь":     6,
		"июне":     6,
		"июля":     7,
		"июль":     7,
		"июле":     7,
		"августа":  8,
		"август":   8,
		"августе":  8,
		"сентября": 9,
		"сентябрь": 9,
		"сентябре": 9,
		"октября":  10,
		"октябрь":  10,
		"октябре":  10,
		"ноября":   11,
		"ноябрь":   11,
		"ноябре":   11,
		"декабря":  12,
		"декабрь":  12,
		"декабре":  12,
	}
}

// getMonthNumber extracts the month number from a string with the month name
func getMonthNumber(monthStr string) int {
	monthStr = strings.ToLower(strings.TrimSpace(monthStr))

	for month, num := range getMonthMap() {
		if strings.Contains(monthStr, month) {
			return num
		}
	}

	return 0
}

// parseNames extracts names from a string and returns them as an array
func parseNames(namesStr string) []string {
	// Split names by comma
	namesSplit := strings.Split(namesStr, ",")

	var cleanNames []string
	for _, name := range namesSplit {
		name = strings.TrimSpace(name)

		// Filter out empty strings and some common phrases
		if name != "" &&
			name != "и иные" &&
			name != "и др." &&
			name != "другие" &&
			!strings.HasPrefix(name, "и ") &&
			!strings.Contains(name, "именины") &&
			!strings.Contains(name, "праздник") &&
			!strings.Contains(name, "день памяти") {
			// Remove possible explanations in parentheses
			if idx := strings.Index(name, "("); idx > 0 {
				name = strings.TrimSpace(name[:idx])
			}
			cleanNames = append(cleanNames, name)
		}
	}

	return cleanNames
}
