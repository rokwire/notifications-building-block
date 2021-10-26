package model

import "time"

// FirebaseToken Firebase token
type FirebaseToken struct {
	Token       string     `json:"token" bson:"token"`
	AppPlatform *string    `json:"app_platform" bson:"app_platform"`
	AppVersion  *string    `json:"app_version" bson:"app_version"`
	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
} // @name FirebaseToken


