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

// InputMessage represents the data structure needed for creating a message. It is the input data for the core module.
type InputMessage struct {
	OrgID string
	AppID string

	ID *string //use ID if given

	Sender                   Sender
	Time                     time.Time
	Priority                 int
	Subject                  string
	Body                     string
	Data                     map[string]string
	InputRecipients          []MessageRecipient
	RecipientsCriteriaList   []RecipientCriteria
	RecipientAccountCriteria map[string]interface{}
	Topic                    *string
	Topics                   []string
}

// InputMessageRecipient represents the data structure needed for creating a message recipient. It is the input data for the core module.
type InputMessageRecipient struct {
	UserID string
	Mute   bool
}

// Message wraps all needed information for the notification. Use either recipients, recipients_criteria or topic in order to address end users
// @Description wraps all needed information for the notification
// @ID Message
type Message struct {
	OrgID string `json:"org_id" bson:"org_id"`
	AppID string `json:"app_id" bson:"app_id"`

	ID       string            `json:"id" bson:"_id"`
	Time     time.Time         `json:"time" bson:"time"`
	Priority int               `json:"priority" bson:"priority"`
	Subject  string            `json:"subject" bson:"subject"`
	Sender   Sender            `json:"sender,omitempty" bson:"sender,omitempty"`
	Body     string            `json:"body" bson:"body"`
	Data     map[string]string `json:"data" bson:"data"`

	//recipients related
	Recipients               []MessageRecipient     `json:"recipients" bson:"recipients"` //keep it for back compatability
	RecipientsCriteriaList   []RecipientCriteria    `json:"recipients_criteria_list" bson:"recipients_criteria_list"`
	RecipientAccountCriteria map[string]interface{} `json:"recipient_account_criteria" bson:"recipient_account_criteria"`
	Topic                    *string                `json:"topic" bson:"topic"`
	Topics                   []string               `json:"topics" bson:"topics"`

	//initialy calculated recipients count
	//if nil then it means that the message was created before the refactoring
	CalculatedRecipientsCount *int `json:"calculated_recipients_count" bson:"calculated_recipients_count"`

	DateCreated *time.Time `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// IsSender checks if the user is a sender
func (m *Message) IsSender(userID string) bool {
	if m.Sender.User != nil && userID == m.Sender.User.UserID {
		return true
	}
	return false
}

// Sender is a system generated fingerprint for the originator of the message. It may be a user from the admin app or an external system
// @name Sender
// @ID Sender
type Sender struct {
	Type string          `json:"type" bson:"type"` // user or system
	User *CoreAccountRef `json:"user,omitempty" bson:"user,omitempty"`
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
	TotalCount   *int64 `json:"total_count" bson:"total_count"`
	Muted        *int64 `json:"muted_count" bson:"muted_count"`
	Unmuted      *int64 `json:"not_muted_count" bson:"not_muted_count"`
	Read         *int64 `json:"read_count" bson:"read_count"`
	Unread       *int64 `json:"not_read_count" bson:"not_read_count"`
	UnreadUnmute *int64 `json:"not_read_not_mute" bson:"not_read_not_mute"`
}

///
