package model

// TokenInfo wraps the input json while registering token
type TokenInfo struct {
	PreviousToken *string `json:"previous_token" bson:"previous_token"`
	Token         *string `json:"token" bson:"token"`
	AppVersion    *string `json:"app_version" bson:"app_version"`
	AppPlatform   *string `json:"app_platform" bson:"app_platform"`
} // @name TokenInfo
