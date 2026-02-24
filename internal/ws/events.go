package ws

import "encoding/json"

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (e Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
