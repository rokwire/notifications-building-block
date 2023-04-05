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

package core

import "notifications/core/model"

func (app *Application) adminGetMessagesStats(orgID string, appID string, adminAccountID string, source string, offset *int64, limit *int64, order *string) (map[int][]interface{}, error) {
	//1. find the messages
	var senderAccountID *string
	if source == "me" {
		senderAccountID = &adminAccountID
	}
	messages, err := app.storage.FindMessagesByParams(orgID, appID, "administrative", senderAccountID, offset, limit, order)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		//empty
		return map[int][]interface{}{}, nil
	}

	//2. get the messages recipients for the messages
	messagesIDs := make([]string, len(messages))
	for i, message := range messages {
		messagesIDs[i] = message.ID
	}
	allMessagesRecipients, err := app.storage.FindMessagesRecipientsByMessages(messagesIDs)
	if err != nil {
		return nil, err
	}

	//3. construct the result
	result := map[int][]interface{}{}
	for i, message := range messages {

		//find the recipients for the message
		messageRecipients := []model.MessageRecipient{}
		for _, recipient := range allMessagesRecipients {
			if recipient.MessageID == message.ID {
				messageRecipients = append(messageRecipients, recipient)
			}
		}

		result[i] = []interface{}{message, messageRecipients}
	}
	return result, nil
}
