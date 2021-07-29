package model

type Message struct {
	Recipients []Recipient `json:"recipients"`
	Subject    string      `json:"subject"`
	Body       string      `json:"body"`
}
