/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, AppVersion 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package model

import "time"

// User represents user entity and all its relationship with firebase tokens and topics
type User struct {
	ID             string          `json:"id" bson:"_id"`
	FirebaseTokens []FirebaseToken `json:"firebase_tokens" bson:"firebase_tokens"`
	UserID         *string         `json:"user_id" bson:"user_id"`
	Topics         []string        `json:"topics" bson:"topics"`
	DateCreated    time.Time       `json:"date_created" bson:"date_created"`
	DateUpdated    time.Time       `json:"date_updated" bson:"date_updated"`
} //@name User

// AddToken adds topic to the list
func (t *User) AddToken(token string) {
	if t.FirebaseTokens == nil {
		t.FirebaseTokens = []FirebaseToken{}
	}
	exists := false
	for _, entry := range t.FirebaseTokens {
		if token == entry.Token {
			exists = true
			break
		}
	}
	if !exists {
		t.FirebaseTokens = append(t.FirebaseTokens, FirebaseToken{
			Token:       token,
			DateCreated: time.Now().UTC(),
		})
	}
}

// RemoveToken removes a topic
func (t *User) RemoveToken(token string) {
	if t.FirebaseTokens != nil {
		tokens := []FirebaseToken{}
		for _, entry := range t.FirebaseTokens {
			if entry.Token != token {
				tokens = append(tokens, entry)
			}
		}
		t.FirebaseTokens = tokens
	}
}

// HasTopic checks if topic already exists
func (t *User) HasTopic(topic string) bool {
	exists := false
	for _, entry := range t.Topics {
		if topic == entry {
			exists = true
			break
		}
	}
	return exists
}

//////////////////////////

// CoreToken represents shibboleth auth entity
type CoreToken struct {
	ExternalID     *string `json:"uid" bson:"uid"`
	AppID          *string `json:"app_id" bson:"app_id"`
	OrganizationID *string `json:"org_id" bson:"org_id"`
	UserID         *string `json:"sub" bson:"sub"`
	Scope          *string `json:"scope" bson:"scope"`
	Permissions    *string `json:"permissions" bson:"permissions"`
	Anonymous      bool    `json:"anonymous" bson:"anonymous"`
} //@name CoreToken

// CoreUserRef user reference that contains ExternalID & Name
type CoreUserRef struct {
	UserID *string `json:"user_id" bson:"user_id"`
	Name   *string `json:"name" bson:"name"`
} //@name CoreUserRef
