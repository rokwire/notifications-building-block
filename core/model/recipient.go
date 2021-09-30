package model

// Recipient represent recipient of a message
type Recipient struct {
	UID  *string `json:"uid" bson:"uid"`
	Name *string `json:"name" bson:"name"`
} //@name Recipient
