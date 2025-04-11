package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type DayMonth struct {
	ts time.Time
}

func NewDayMonth(ts time.Time) DayMonth {
	return DayMonth{ts: ts}
}

func (d DayMonth) String() string {
	return fmt.Sprintf("%02d%02d", d.ts.Month(), d.ts.Day())
}

// MarshalJSON implements the json.Marshaler interface
func (d DayMonth) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *DayMonth) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if len(s) != 4 {
		return fmt.Errorf("invalid date format: %s, expected MMDD", s)
	}

	month, err := strconv.Atoi(s[:2])
	if err != nil {
		return fmt.Errorf("invalid month: %s", s[:2])
	}

	day, err := strconv.Atoi(s[2:])
	if err != nil {
		return fmt.Errorf("invalid day: %s", s[2:])
	}

	// Use the current year, as we only need the month and day
	d.ts = time.Date(time.Now().Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return nil
}

type NamedaysData struct {
	Date  DayMonth `json:"date"`
	Names []string `json:"names"`
}

type NamedaysDataList []NamedaysData

func (l NamedaysDataList) Len() int {
	return len(l)
}
