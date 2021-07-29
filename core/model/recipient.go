package model

type Recipient struct {
	Uin   *string `json:"uin" bson:"uin"`
	Email *string `json:"email" bson:"email"`
	Phone *string `json:"phone" bson:"phone"`
	Topic *string `json:"topic" bson:"topic"`
}
