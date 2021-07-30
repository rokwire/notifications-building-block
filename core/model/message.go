package model

type Message struct {
	Recipients []Recipient `json:"recipients"`
	Topic      *string     `json:"topic"`
	Subject    string      `json:"subject"`
	Body       string      `json:"body"`
}
