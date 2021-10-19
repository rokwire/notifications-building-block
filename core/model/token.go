package model

import "time"

// FirebaseToken Firebase token
type FirebaseToken struct {
	Token       string    `json:"token" bson:"token"`
	Platform    *string   `json:"platform" bson:"platform"`
	Version     *string   `json:"version" bson:"version"`
	DateCreated time.Time `json:"date_created" bson:"date_created"`
	DateUpdated time.Time `json:"date_updated" bson:"date_updated"`
} //@name FirebaseToken

type TokenInfo struct {
	PreviousToken *string `json:"previous_token" bson:"previous_token"`
	Token         *string `json:"token" bson:"token"`
	Platform      *string `json:"platform" bson:"platform"`
	Version       *string `json:"version" bson:"version"`
} // @name TokenInfo
