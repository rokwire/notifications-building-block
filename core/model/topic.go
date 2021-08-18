package model

import "time"

// Topic wraps a firebase topic and description
type Topic struct {
	Name        string    `json:"name" bson:"_id"`
	UserIDs     []string  `json:"user_ids" bson:"user_ids"`
	Description *string   `json:"description" bson:"description"`
	DateCreated time.Time `json:"date_created" bson:"date_created"`
	DateUpdated time.Time `json:"date_updated" bson:"date_updated"`
} //@name Topic

// AddUserID adds user id to the list
func (t *Topic) AddUserID(userID string) {
	if t.UserIDs == nil {
		t.UserIDs = []string{}
	}
	exists := false
	for _, entry := range t.UserIDs {
		if userID == entry {
			exists = true
			break
		}
	}
	if !exists {
		t.UserIDs = append(t.UserIDs, userID)
	}
}

// RemoveUserID removes a topic
func (t *Topic) RemoveUserID(userID string) {
	if t.UserIDs != nil {
		userIDs := []string{}
		for _, entry := range t.UserIDs {
			if entry != userID {
				userIDs = append(userIDs, entry)
			}
		}
		t.UserIDs = userIDs
	}
}
