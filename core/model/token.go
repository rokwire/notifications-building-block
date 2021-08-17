package model

import "time"

// FirebaseTokenMapping mapped token to a Shibboleth user
type FirebaseTokenMapping struct {
	ID          string    `json:"id" bson:"_id"`
	Tokens      []string  `json:"firebase_tokens" bson:"firebase_tokens"`
	UserID      *string   `json:"user_id" bson:"user_id"`
	Topics      []string  `json:"topics" bson:"topics"`
	DateCreated time.Time `json:"date_created" bson:"date_created"`
	DateUpdated time.Time `json:"date_updated" bson:"date_updated"`
} //@name FirebaseTokenMapping

// AddTopic adds topic to the list
func (t *FirebaseTokenMapping) AddTopic(topic string) {
	if t.Topics == nil {
		t.Topics = []string{}
	}
	exists := false
	for _, entry := range t.Topics {
		if topic == entry {
			exists = true
			break
		}
	}
	if !exists {
		t.Topics = append(t.Topics, topic)
	}
}

// RemoveTopic removes a topic
func (t *FirebaseTokenMapping) RemoveTopic(topic string) {
	if t.Topics != nil {
		topics := []string{}
		for _, entry := range t.Topics {
			if entry != topic {
				topics = append(topics, entry)
			}
		}
		t.Topics = topics
	}
}


