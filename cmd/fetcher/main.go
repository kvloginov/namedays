package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kvloginov/namedays/internal/domain"
	fetch "github.com/kvloginov/namedays/internal/fetch"
)

func main() {
	sourceType := flag.String("source", "krestilnoe", "The source to fetch namedays from (krestilnoe, calend, pravmir, or merge)")
	flag.Parse()

	var fetcher fetch.Fetcher
	var filename string
	var namedays []domain.NamedaysData
	var err error

	switch *sourceType {
	case "krestilnoe":
		fetcher = fetch.NewKrestilnoeFetcher()
		filename = "data/krestilnoe_namedays.json"
		namedays, err = fetcher.FetchAllNamedays()
		if err != nil {
			log.Fatalf("error fetching namedays: %v", err)
		}
	case "calend":
		fetcher = fetch.NewCalendFetcher()
		filename = "data/calend_namedays.json"
		namedays, err = fetcher.FetchAllNamedays()
		if err != nil {
			log.Fatalf("error fetching namedays: %v", err)
		}
	case "pravmir":
		fetcher = fetch.NewPravmirFetcher()
		filename = "data/pravmir_namedays.json"
		namedays, err = fetcher.FetchAllNamedays()
		if err != nil {
			log.Fatalf("error fetching namedays: %v", err)
		}
	case "merge":
		filename = "data/merged_namedays.json"
		namedays, err = mergeNamedaysFiles()
		if err != nil {
			log.Fatalf("error merging namedays: %v", err)
		}
	default:
		log.Fatalf("unknown source type: %s", *sourceType)
	}

	// save to file
	jsonData, err := json.Marshal(namedays)
	if err != nil {
		log.Fatalf("error marshalling namedays: %v", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		log.Fatalf("error writing file: %v", err)
	}

	fmt.Printf("Successfully fetched namedays from %s and saved to %s\n", *sourceType, filename)
}

func mergeNamedaysFiles() ([]domain.NamedaysData, error) {
	// Map to store merged namedays data by date
	mergedMap := make(map[string]map[string]bool)

	// Find all namedays files in data directory
	files, err := filepath.Glob("data/*_namedays.json")
	if err != nil {
		return nil, fmt.Errorf("error finding namedays files: %v", err)
	}

	for _, file := range files {
		// Skip merged_namedays.json if it exists
		if strings.Contains(file, "merged_namedays.json") {
			continue
		}

		// Read file
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", file, err)
		}

		// Parse JSON
		var namedaysList []domain.NamedaysData
		if err := json.Unmarshal(data, &namedaysList); err != nil {
			return nil, fmt.Errorf("error unmarshalling file %s: %v", file, err)
		}

		// Merge data
		for _, nameday := range namedaysList {
			date := nameday.Date.String()

			// Initialize map for this date if not exists
			if _, ok := mergedMap[date]; !ok {
				mergedMap[date] = make(map[string]bool)
			}

			// Add names for this date
			for _, name := range nameday.Names {
				mergedMap[date][name] = true
			}
		}
	}

	// Convert merged map back to NamedaysData slice
	var result []domain.NamedaysData
	var dates []string
	for dateStr := range mergedMap {
		dates = append(dates, dateStr)
	}

	// Sort dates
	sort.Strings(dates)

	for _, dateStr := range dates {
		// Parse date string back to DayMonth
		month, _ := strconv.Atoi(dateStr[:2])
		day, _ := strconv.Atoi(dateStr[2:])
		date := domain.NewDayMonth(time.Date(time.Now().Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC))

		// Get unique names and sort them
		var names []string
		for name := range mergedMap[dateStr] {
			names = append(names, name)
		}
		sort.Strings(names)

		// Add to result
		result = append(result, domain.NamedaysData{
			Date:  date,
			Names: names,
		})
	}

	return result, nil
}
