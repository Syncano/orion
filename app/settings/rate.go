package settings

import "time"

type RateData struct {
	Limit    int64
	Duration time.Duration
}
