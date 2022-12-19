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

import "time"

// QueueItem represent notifications queue data item
type QueueItem struct {
	OrgID string `bson:"org_id"`
	AppID string `bson:"app_id"`
	ID    string `bson:"_id"`

	//who to send
	MessageRecipientID string `bson:"message_recipient_id"`
	UserID             string `bson:"user_id"`

	//what to send
	Subject string            `bson:"subject"`
	Body    string            `bson:"body"`
	Data    map[string]string `bson:"data"`

	//when to send
	Time     time.Time `bson:"time"`
	Priority int       `bson:"priority"`
}
