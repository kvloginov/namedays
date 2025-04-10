package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kvloginov/namedays/internal/parser"
)

func main() {
	log.Println("Starting namedays fetcher...")

	fetcher := parser.NewNamedaysFetcher()

	// Map to store namedays with date as key (MMDD format)
	namedaysMap := make(map[string][]string)

	// Current date
	now := time.Now()
	currentYear := now.Year()

	// Start date (January 1st of current year)
	startDate := time.Date(currentYear, time.January, 1, 0, 0, 0, 0, time.Local)

	// Process each day of the year
	var successCount, errorCount int

	for day := 0; day < 366; day++ { // 366 to handle leap years
		date := startDate.AddDate(0, 0, day)

		// If we've moved to next year, break
		if date.Year() > currentYear {
			break
		}

		// Format date as MMDD
		dateKey := parser.FormatDateKey(date)

		log.Printf("Fetching namedays for %s (%s)...", date.Format("2006-01-02"), dateKey)

		// Fetch namedays for the date
		names, err := fetcher.FetchNamedays(date)
		if err != nil {
			log.Printf("Error fetching namedays for %s: %v", date.Format("2006-01-02"), err)
			errorCount++
			continue
		}

		// Add to map
		namedaysMap[dateKey] = names

		log.Printf("Got %d names for %s: %v", len(names), dateKey, names)
		successCount++

		// Sleep to avoid overwhelming the server
		time.Sleep(500 * time.Millisecond)
	}

	// Save to JSON file
	jsonData, err := json.MarshalIndent(namedaysMap, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling namedays map: %v", err)
	}

	outputFile := "namedays.json"
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		log.Fatalf("Error writing to file %s: %v", outputFile, err)
	}

	log.Printf("Done! Processed %d days successfully, %d errors. Data saved to %s",
		successCount, errorCount, outputFile)

	// Print statistics
	fmt.Printf("\nNamedays Statistics:\n")
	fmt.Printf("Total days processed: %d\n", successCount+errorCount)
	fmt.Printf("Successful days: %d\n", successCount)
	fmt.Printf("Failed days: %d\n", errorCount)
	fmt.Printf("Total unique names: %d\n", countUniqueNames(namedaysMap))
	fmt.Printf("Output file: %s\n", outputFile)
}

// countUniqueNames counts unique names in the namedays map
func countUniqueNames(namedaysMap map[string][]string) int {
	uniqueNames := make(map[string]bool)

	for _, names := range namedaysMap {
		for _, name := range names {
			uniqueNames[name] = true
		}
	}

	return len(uniqueNames)
}
