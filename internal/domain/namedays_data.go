package domain

import (
	"fmt"
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

type NamedaysData struct {
	Date  DayMonth `json:"date"`
	Names []string `json:"names"`
}

type NamedaysDataList []NamedaysData

func (l NamedaysDataList) Len() int {
	return len(l)
}
