package main

import (
	"encoding/json"
	"log"
	"os"

	fetch "github.com/kvloginov/namedays/internal/fetch"
)

func main() {
	krestilnoeFetcher := fetch.NewKrestilnoeFetcher()

	krestilnoeNamedays, err := krestilnoeFetcher.FetchAllNamedays()
	if err != nil {
		log.Fatalf("error fetching namedays: %v", err)
	}

	// save to file
	jsonData, err := json.Marshal(krestilnoeNamedays)
	if err != nil {
		log.Fatalf("error marshalling namedays: %v", err)
	}
	os.WriteFile("data/krestilnoe_namedays.json", jsonData, 0644)
}
