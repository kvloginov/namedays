package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	fetch "github.com/kvloginov/namedays/internal/fetch"
)

func main() {
	sourceType := flag.String("source", "krestilnoe", "The source to fetch namedays from (krestilnoe, calend, or pravmir)")
	flag.Parse()

	var fetcher fetch.Fetcher
	var filename string

	switch *sourceType {
	case "krestilnoe":
		fetcher = fetch.NewKrestilnoeFetcher()
		filename = "data/krestilnoe_namedays.json"
	case "calend":
		fetcher = fetch.NewCalendFetcher()
		filename = "data/calend_namedays.json"
	case "pravmir":
		fetcher = fetch.NewPravmirFetcher()
		filename = "data/pravmir_namedays.json"
	default:
		log.Fatalf("unknown source type: %s", *sourceType)
	}

	namedays, err := fetcher.FetchAllNamedays()
	if err != nil {
		log.Fatalf("error fetching namedays: %v", err)
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
