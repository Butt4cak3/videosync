package internal

import (
	"encoding/json"
	"fmt"
)

type MessageType string

const (
	Init    MessageType = "init"
	Play    MessageType = "play"
	Pause   MessageType = "pause"
	Load    MessageType = "load"
	LoadUrl MessageType = "loadurl"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload any         `json:"payload"`
}

type InitMessage struct {
	VideoId       string  `json:"videoId"`
	VideoPos      float32 `json:"videoPos"`
	PlaybackState int     `json:"playbackState"`
}

type PlayMessage struct {
	Position float32 `json:"position"`
}

type PauseMessage struct {
	Position float32 `json:"position"`
}

type LoadMessage struct {
	VideoId string `json:"videoId"`
}

type LoadUrlMessage struct {
	Url string `json:"url"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	var temp struct {
		Type    MessageType     `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	m.Type = temp.Type

	switch temp.Type {
	case Play:
		var payload PlayMessage
		if err := json.Unmarshal(temp.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	case Pause:
		var payload PauseMessage
		if err := json.Unmarshal(temp.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	case Load:
		var payload LoadMessage
		if err := json.Unmarshal(temp.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	case LoadUrl:
		var payload LoadUrlMessage
		if err := json.Unmarshal(temp.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	default:
		return fmt.Errorf("unknown message type: %s", temp.Type)
	}

	return nil
}
