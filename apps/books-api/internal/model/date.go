package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

var dateLayouts = []string{
	"2006-01-02",
	"02-01-2006",
	"2006/01/02",
	"January 2, 2006",
	"Jan 2, 2006",
	time.RFC3339,
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid date format (string expected): %w", err)
	}

	if s == "" {
		d.Time = time.Time{}
		return nil
	}

	for _, layout := range dateLayouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			d.Time = t
			return nil
		}
	}

	return fmt.Errorf("cannot parse date: %s", s)
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte(`null`), nil
	}

	s := d.Time.Format("2006-01-02")
	return json.Marshal(s)
}
