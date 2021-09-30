package model

import "time"

// FirebaseToken Firebase token
type FirebaseToken struct {
	Token       string    `json:"token" bson:"token"`
	DateCreated time.Time `json:"date_created" bson:"date_created"`
} //@name FirebaseToken
