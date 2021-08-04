package model

import "time"

// FirebaseTokenMapping mapped token to a Shibboleth user
type FirebaseTokenMapping struct {
	Token       string    `json:"firebase_token" bson:"_id"`
	DeviceID    *string   `json:"device_id" bson:"device_id"`
	Uin         *string   `json:"uin" bson:"uin"`
	Email       *string   `json:"email" bson:"email"`
	Phone       *string   `json:"phone" bson:"phone"`
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
