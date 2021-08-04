package model

import (
	"time"
)

// Message wraps all needed information for the notification
// @Description wraps all needed information for the notification
// @ID Message
type Message struct {
	ID          *string     `json:"id" bson:"_id"`
	DateCreated *time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time  `json:"date_updated" bson:"date_updated"`
	DateSent    *time.Time  `json:"date_sent" bson:"date_sent"`
	Sent        bool        `json:"sent" bson:"sent"`
	Recipients  []Recipient `json:"recipients" bson:"recipients"`
	Topic       *string     `json:"topic" bson:"topic"`
	Subject     string      `json:"subject" bson:"subject"`
	Sender      *Sender     `json:"sender,omitempty" bson:"sender,omitempty"`
	Body        string      `json:"body" bson:"body"`
}

// Sender is a system generated fingerprint for the originator of the message. It may be a user from the admin app or an external system
// @name Sender
// @ID Sender
type Sender struct {
	Type string `json:"type" bson:"type"` // user or system
	User *User  `json:"user,omitempty" bson:"user,omitempty"`
}
