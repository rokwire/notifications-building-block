// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"time"
)

// Message wraps all needed information for the notification. Use either recipients, recipients_criteria or topic in order to address end users
// @Description wraps all needed information for the notification
// @ID Message
type Message struct {
	OrgID string `json:"org_id" bson:"org_id"`
	AppID string `json:"app_id" bson:"app_id"`

	ID       string            `json:"id" bson:"_id"`
	Priority int               `json:"priority" bson:"priority"`
	Subject  string            `json:"subject" bson:"subject"`
	Sender   Sender            `json:"sender,omitempty" bson:"sender,omitempty"`
	Body     string            `json:"body" bson:"body"`
	Data     map[string]string `json:"data" bson:"data"`

	//recipients related
	RecipientsCriteriaList []RecipientCriteria `json:"recipients_criteria_list" bson:"recipients_criteria_list"`
	Topic                  *string             `json:"topic" bson:"topic"`

	DateCreated *time.Time `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// HasUser checks if the user is the sender or as a recipient for the current message
// Use better name
func (m *Message) HasUser(id string) bool {
	for _, recipient := range m.Recipients {
		if recipient.UserID == id {
			return true
		}
	}

	if m.Sender.User != nil && id == m.Sender.User.UserID {
		return true
	}
	return false
}

// Sender is a system generated fingerprint for the originator of the message. It may be a user from the admin app or an external system
// @name Sender
// @ID Sender
type Sender struct {
	Type string       `json:"type" bson:"type"` // user or system
	User *CoreUserRef `json:"user,omitempty" bson:"user,omitempty"`
}

// RecipientCriteria defines common search criteria for end users and their FCM tokens
// @name RecipientCriteria
// @ID RecipientCriteria
type RecipientCriteria struct {
	AppVersion  *string `json:"app_version" bson:"app_version"`
	AppPlatform *string `json:"app_platform" bson:"app_platform"`
}

// MessagesStats wraps messages statistics aggregation result
// @name MessagesStats
// @ID MessagesStats
type MessagesStats struct {
	TotalCount *int64 `json:"total_count" bson:"total_count"`
	Muted      *int64 `json:"muted_count" bson:"muted_count"`
	Unmuted    *int64 `json:"not_muted_count" bson:"not_muted_count"`
	Read       *int64 `json:"read_count" bson:"read_count"`
	Unread     *int64 `json:"not_read_count" bson:"not_read_count"`
}

///

//InputMessage is passed by the adapters for creating a message in the core module
type InputMessage struct {
	OrgID string `json:"org_id"`
	AppID string `json:"app_id"`

	Priority int               `json:"priority"`
	Subject  string            `json:"subject"`
	Sender   *InputSender      `json:"sender"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data"`

	//recipients related
	Recipients             []InputMessageRecipient  `json:"recipients"`
	RecipientsCriteriaList []InputRecipientCriteria `json:"recipients_criteria_list"`
	Topic                  *string                  `json:"topic"`
}

// InputSender is passed by the adapters for creating a message in the core module
type InputSender struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

// InputMessageRecipient is passed by the adapters for creating a message in the core module
type InputMessageRecipient struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Mute   bool   `json:"mute"`
}

// InputRecipientCriteria is passed by the adapters for creating a message in the core module
type InputRecipientCriteria struct {
	AppVersion  *string `json:"app_version"`
	AppPlatform *string `json:"app_platform"`
}
