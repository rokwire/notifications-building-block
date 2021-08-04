package model

// Recipient represent recipient of a message
type Recipient struct {
	Uin   *string `json:"uin" bson:"uin"`
	Email *string `json:"email" bson:"email"`
	Phone *string `json:"phone" bson:"phone"`
} //@name Recipient
