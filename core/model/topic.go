package model

import "time"

// Topic wraps a firebase topic and description
type Topic struct {
	Name        string    `json:"name" bson:"_id"`
	AppID       *string   `json:"app_id" bson:"app_id"`
	OrgID       *string   `json:"org_id" bson:"org_id"`
	Description *string   `json:"description" bson:"description"`
	DateCreated time.Time `json:"date_created" bson:"date_created"`
	DateUpdated time.Time `json:"date_updated" bson:"date_updated"`
} // @name Topic
