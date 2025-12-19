package internal

import "time"

type PlaybackState int

const (
	Unstarted PlaybackState = -1
	Ended     PlaybackState = 0
	Playing   PlaybackState = 1
	Paused    PlaybackState = 2
)

type Playback struct {
	VideoId            string
	State              PlaybackState
	LatestPosition     float32
	LatestPositionTime time.Time
}

func (p Playback) Position() float32 {
	if p.State != Playing {
		return p.LatestPosition
	}

	elapsed := time.Since(p.LatestPositionTime)
	return p.LatestPosition + float32(elapsed.Seconds())
}
