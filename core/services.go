/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package core

import (
	"fmt"
	"notifications/core/model"
)

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) storeFirebaseToken(tokenInfo *model.TokenInfo, user *model.CoreToken) error {
	return app.storage.StoreFirebaseToken(tokenInfo, user)
}

func (app *Application) subscribeToTopic(token string, user *model.CoreToken, topic string) error {
	var err error
	if user != nil {
		err = app.storage.SubscribeToTopic(token, user.UserID, topic)
		if err == nil {
			err = app.firebase.SubscribeToTopic(token, topic)
		}
	} else if token != "" {
		// Treat this user as anonymous.
		err = app.firebase.SubscribeToTopic(token, topic)
	}
	return err
}

func (app *Application) unsubscribeToTopic(token string, user *model.CoreToken, topic string) error {
	var err error
	if user != nil {
		err = app.storage.UnsubscribeToTopic(token, user.UserID, topic)
		if err == nil {
			err = app.firebase.UnsubscribeToTopic(token, topic)
		}
	} else if token != "" {
		// Treat this user as anonymous.
		err = app.firebase.UnsubscribeToTopic(token, topic)
	}
	return err
}

func (app *Application) getTopics() ([]model.Topic, error) {
	return app.storage.GetTopics()
}

func (app *Application) appendTopic(topic *model.Topic) (*model.Topic, error) {
	return app.storage.InsertTopic(topic)
}

func (app *Application) updateTopic(topic *model.Topic) (*model.Topic, error) {
	return app.storage.UpdateTopic(topic)
}

func (app *Application) createMessage(user *model.CoreToken, message *model.Message) (*model.Message, error) {
	var persistedMessage *model.Message
	var err error
	if message.ID != nil {
		return nil, fmt.Errorf("message with id (%s): is already sent", *message.ID)
	}

	if user != nil {
		message.Sender = &model.Sender{Type: "token", User: &model.CoreUserRef{UserID: user.UserID}}
	} else {
		message.Sender = &model.Sender{Type: "system"}
	}
	storeInInbox := len(message.Subject) > 0 && len(message.Body) > 0
	if storeInInbox {
		persistedMessage, err = app.storage.CreateMessage(message)
	}

	hasRecipients := len(message.Recipients) > 0
	if hasRecipients{
		tokens, err := app.storage.GetFirebaseTokensByRecipients(message.Recipients, message.Topic)
		if err != nil {
			return nil, err
		}
		if len(tokens) > 0 {
			for _, token := range tokens {
				sendErr := app.firebase.SendNotificationToToken(token, message.Subject, message.Body, message.Data)
				if sendErr != nil {
					fmt.Printf("error send notification to token (%s): %s", token, err)
				}
			}
		}
	} else if message.Topic != nil {
		recipients, err := app.storage.GetRecipientsByTopic(*message.Topic)
		if err != nil {
			fmt.Printf("error retrieving recipients by topic (%s): %s", *message.Topic, err)
		}
		if recipients != nil {
			message.Recipients = recipients
			if storeInInbox {
				persistedMessage, err = app.storage.UpdateMessage(message) // just update the message
				if err != nil {
					fmt.Printf("error storing the message to topic %s: %s", *message.Topic, err)
				}
			}
		}

		err = app.firebase.SendNotificationToTopic(*message.Topic, message.Subject, message.Body, message.Data)
		if err != nil {
			fmt.Printf("error send notification to topic (%s): %s", *message.Topic, err)
		}
	} else if message.RecipientsCriteriaList != nil {
		recipients, err := app.storage.GetRecipientsByRecipientCriterias(message.RecipientsCriteriaList)
		if err != nil {
			fmt.Printf("error retrieving recipients by topic (%s): %s", *message.Topic, err)
		}
		if recipients != nil {
			message.Recipients = recipients
			if storeInInbox {
				persistedMessage, err = app.storage.UpdateMessage(message) // just update the message
				if err != nil {
					fmt.Printf("error storing the message to topic %s: %s", *message.Topic, err)
				}
			}
		}

		tokens, err := app.storage.GetFirebaseTokensByRecipients(message.Recipients, nil)
		if err != nil {
			return nil, err
		}
		if len(tokens) > 0 {
			for _, token := range tokens {
				sendErr := app.firebase.SendNotificationToToken(token, message.Subject, message.Body, message.Data)
				if sendErr != nil {
					fmt.Printf("error send notification to token (%s): %s", token, err)
				}
			}
		}
	}

	if err != nil {
		return nil, err
	}

	return persistedMessage, err
}

func (app *Application) getMessages(userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	return app.storage.GetMessages(userID, messageIDs, startDateEpoch, endDateEpoch, filterTopic, offset, limit, order)
}

func (app *Application) getMessage(ID string) (*model.Message, error) {
	return app.storage.GetMessage(ID)
}

func (app *Application) updateMessage(user *model.CoreToken, message *model.Message) (*model.Message, error) {
	if message.ID != nil {
		persistedMessage, err := app.storage.GetMessage(*message.ID)
		if err == nil && persistedMessage != nil {
			if persistedMessage.Sender.User != nil && persistedMessage.Sender.User.UserID == user.UserID {
				return app.storage.UpdateMessage(message)
			}
			return nil, fmt.Errorf("only creator can update the original message")
		}
	}
	return nil, fmt.Errorf("missing id or record")
}

func (app *Application) deleteUserMessage(user *model.CoreToken, messageID string) error {
	return app.storage.DeleteUserMessage(*user.UserID, messageID)
}

func (app *Application) deleteMessage(ID string) error {
	return app.storage.DeleteMessage(ID)
}

func (app *Application) getAllAppVersions() ([]model.AppVersion, error) {
	return app.storage.GetAllAppVersions()
}

func (app *Application) getAllAppPlatforms() ([]model.AppPlatform, error) {
	return app.storage.GetAllAppPlatforms()
}

func (app *Application) findUserByID(userID string) (*model.User, error) {
	return app.storage.FindUserByID(userID)
}
