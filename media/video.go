package media

import "time"

type Video struct {
	Id       string        `json:"id"`
	Title    string        `json:"title"`
	Duration time.Duration `json:"duration"`
}
