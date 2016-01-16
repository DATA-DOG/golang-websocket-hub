package hub

import "encoding/json"

type Message struct {
	Data  interface{} `json:"data"`
	Topic string      `json:"topic"`
}

func (m *Message) bytes() ([]byte, error) {
	return json.Marshal(m)
}

type MailMessage struct {
	Message  *Message
	Username string
}
