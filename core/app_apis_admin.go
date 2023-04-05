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

func (app *Application) adminGetMessagesStats(orgID string, appID string, adminAccountID string, source string, offset *int64, limit *int64, order *string) (map[string][]interface{}, error) {
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
		return map[string][]interface{}{}, nil
	}

	//context.Context
	//FindMessagesByParams(ctx context.Context, orgID string, appID string, senderType string, senderAccountID *string, offset *int64, limit *int64, order *string) ([]model.Message, error)

	/*now := time.Now()

	message1 := model.Message{ID: "1", DateCreated: &now, Time: now, Body: "Body 1", Sender: model.Sender{Type: "administrative", User: &model.CoreAccountRef{UserID: "100", Name: "Ime 1"}}}
	recps1 := []model.MessageRecipient{{ID: "1", Read: true}} //do not put nil
	sect1 := []interface{}{message1, recps1}

	message2 := model.Message{ID: "2", DateCreated: &now, Time: now, Body: "Body 2", Sender: model.Sender{Type: "administrative", User: &model.CoreAccountRef{UserID: "200", Name: "Ime 2"}}}
	recps2 := []model.MessageRecipient{{ID: "2", Read: false}}
	sect2 := []interface{}{message2, recps2} //do not put nil
	*/

	result := map[string][]interface{}{}
	//result[message1.ID] = sect1
	//result[message2.ID] = sect2

	//TODO do not return nil
	return result, nil
}
