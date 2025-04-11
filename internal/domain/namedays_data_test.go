package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDayMonthJSONMarshaling(t *testing.T) {
	// Create a date of April 15, 2023
	date := time.Date(2023, time.April, 15, 0, 0, 0, 0, time.UTC)
	dm := NewDayMonth(date)

	// Check the String() method
	if dm.String() != "0415" {
		t.Errorf("Expected '0415', got '%s'", dm.String())
	}

	// Create a structure with our DayMonth
	data := NamedaysData{
		Date:  dm,
		Names: []string{"John", "Jane"},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Check that the JSON contains the expected string for the date
	expectedJSON := `{"date":"0415","names":["John","Jane"]}`
	actualJSON := string(jsonBytes)
	if actualJSON != expectedJSON {
		t.Errorf("Expected JSON:\n%s\nGot:\n%s", expectedJSON, actualJSON)
	}

	// Check the deserialization
	var unmarshaled NamedaysData
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that after deserialization the date matches the expectations
	if unmarshaled.Date.String() != "0415" {
		t.Errorf("Expected date '0415', got '%s'", unmarshaled.Date.String())
	}

	// Check that the names also match the expectations
	if len(unmarshaled.Names) != 2 || unmarshaled.Names[0] != "John" || unmarshaled.Names[1] != "Jane" {
		t.Errorf("Names don't match: expected [John Jane], got %v", unmarshaled.Names)
	}
}
