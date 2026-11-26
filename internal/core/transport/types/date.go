package types

import (
	"encoding/json"
	"time"
)

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(time.DateOnly))
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}

	d.Time = t

	return nil
}

func (d *Date) ToDomain() time.Time {
	return d.Time
}
