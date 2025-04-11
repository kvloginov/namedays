package main

import (
	"encoding/json"
	"log"
	"os"

	fetch "github.com/kvloginov/namedays/internal/fetch"
)

func main() {
	calendFetcher := fetch.NewCalendFetcher()

	calendNamedays, err := calendFetcher.FetchAllNamedays()
	if err != nil {
		log.Fatalf("error fetching namedays: %v", err)
	}

	// save to file
	jsonData, err := json.Marshal(calendNamedays)
	if err != nil {
		log.Fatalf("error marshalling namedays: %v", err)
	}
	os.WriteFile("data/calend_namedays.json", jsonData, 0644)
}
