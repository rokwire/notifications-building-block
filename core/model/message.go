package model

import (
	"time"
)

// Message wraps all needed information for the notification
// @Description wraps all needed information for the notification
// @ID Message
type Message struct {
	ID          *string           `json:"id" bson:"_id"`
	DateCreated *time.Time        `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time        `json:"date_updated" bson:"date_updated"`
	Priority    int               `json:"priority" bson:"priority"`
	Recipients  []Recipient       `json:"recipients" bson:"recipients"`
	Topic       *string           `json:"topic" bson:"topic"`
	Subject     string            `json:"subject" bson:"subject"`
	Sender      *Sender           `json:"sender,omitempty" bson:"sender,omitempty"`
	Body        string            `json:"body" bson:"body"`
	Data        map[string]string `json:"data" bson:"data"`
}

// HasUser checks if the user is the sender or as a recipient for the current message
// Use better name
func (m *Message) HasUser(token *CoreToken) bool {
	if token != nil {
		for _, recipient := range m.Recipients {
			if recipient.UserID == token.UserID {
				return true
			}
		}

		if m.Sender.User.UserID != nil && token.UserID == m.Sender.User.UserID {
			return true
		}
	}
	return false
}

// Sender is a system generated fingerprint for the originator of the message. It may be a user from the admin app or an external system
// @name Sender
// @ID Sender
type Sender struct {
	Type string       `json:"type" bson:"type"` // user or system
	User *CoreUserRef `json:"user,omitempty" bson:"user,omitempty"`
}
