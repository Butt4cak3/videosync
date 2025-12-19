package internal

import (
	"encoding/json"
	"fmt"
)

type MessageType string

const (
	Init  MessageType = "init"
	Play  MessageType = "play"
	Pause MessageType = "pause"
	Load  MessageType = "load"
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
		var playPayload PlayMessage
		if err := json.Unmarshal(temp.Payload, &playPayload); err != nil {
			return err
		}
		m.Payload = playPayload
	case Pause:
		var pausePayload PauseMessage
		if err := json.Unmarshal(temp.Payload, &pausePayload); err != nil {
			return err
		}
		m.Payload = pausePayload
	case Load:
		var loadPayload LoadMessage
		if err := json.Unmarshal(temp.Payload, &loadPayload); err != nil {
			return err
		}
		m.Payload = loadPayload
	default:
		return fmt.Errorf("unknown messaget ype: %s", temp.Type)
	}

	return nil
}
