package model

// AppVersion wraps app version number
type AppVersion struct {
	Name *string `json:"name" bson:"name"`
} //@name AppVersion

// AppPlatform wraps app platform name
type AppPlatform struct {
	Name *string `json:"name" bson:"name"`
} //@name AppPlatform
