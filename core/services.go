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

import (
	"notifications/core/model"
)

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) storeFirebaseToken(tokenInfo *model.TokenInfo, user *model.CoreToken) error {
	//return app.storage.StoreFirebaseToken(tokenInfo, user)
	return nil
}

func (app *Application) subscribeToTopic(token string, user *model.CoreToken, topic string) error {
	var err error
	/*if user != nil {
		err = app.storage.SubscribeToTopic(token, user.UserID, topic)
		if err == nil {
			err = app.firebase.SubscribeToTopic(token, topic)
		}
	} else if token != "" {
		// Treat this user as anonymous.
		err = app.firebase.SubscribeToTopic(token, topic)
	} */
	return err
}

func (app *Application) unsubscribeToTopic(token string, user *model.CoreToken, topic string) error {
	var err error
	/*if user != nil {
		err = app.storage.UnsubscribeToTopic(token, user.UserID, topic)
		if err == nil {
			err = app.firebase.UnsubscribeToTopic(token, topic)
		}
	} else if token != "" {
		// Treat this user as anonymous.
		err = app.firebase.UnsubscribeToTopic(token, topic)
	}*/
	return err
}

func (app *Application) getTopics() ([]model.Topic, error) {
	return nil, nil
	//return app.storage.GetTopics()
}

func (app *Application) appendTopic(topic *model.Topic) (*model.Topic, error) {
	return app.storage.InsertTopic(topic)
}

func (app *Application) updateTopic(topic *model.Topic) (*model.Topic, error) {
	return app.storage.UpdateTopic(topic)
}

func (app *Application) createMessage(user *model.CoreToken, message *model.Message, async bool) (*model.Message, error) {
	return nil, nil
	/*	var persistedMessage *model.Message
		var err error
		if message.ID != nil {
			return nil, fmt.Errorf("message with id (%s): is already sent", *message.ID)
		}

		if user != nil {
			message.Sender = &model.Sender{Type: "user", User: &model.CoreUserRef{UserID: user.UserID, Name: user.Name}}
		} else {
			message.Sender = &model.Sender{Type: "system"}
		}
		storeInInbox := len(message.Subject) > 0 || len(message.Body) > 0
		if storeInInbox {
			persistedMessage, err = app.storage.CreateMessage(message)
			if err != nil {
				fmt.Printf("error on creating a message: %s", err)
				return nil, fmt.Errorf("error on creating a message: %s", err)
			}
			log.Printf("message %s has been created", *persistedMessage.ID)
		}

		messageRecipients := []model.Recipient{}
		checkCriteria := true

		// recipients from message
		if len(message.Recipients) > 0 {
			messageRecipients = append(messageRecipients, message.Recipients...)
		}

		// recipients from topic
		if message.Topic != nil {
			topicRecipients, err := app.storage.GetRecipientsByTopic(*message.Topic)
			if err != nil {
				fmt.Printf("error retrieving recipients by topic (%s): %s", *message.Topic, err)
			} else {
				log.Printf("retrieve recipients (%+v) for topic (%s)", topicRecipients, *message.Topic)
			}

			if len(topicRecipients) > 0 {
				if len(messageRecipients) > 0 {
					messageRecipients = getCommonRecipients(messageRecipients, topicRecipients)
				} else {
					messageRecipients = append(messageRecipients, topicRecipients...)
				}
			} else {
				checkCriteria = false
				messageRecipients = nil
			}

			log.Printf("construct recipients (%+v) for message (%s:%s:%s)", messageRecipients, *message.ID, message.Subject, message.Body)
		}

		// recipients from criteria
		if (message.RecipientsCriteriaList != nil) && checkCriteria {
			criteriaRecipients, err := app.storage.GetRecipientsByRecipientCriterias(message.RecipientsCriteriaList)
			if err != nil {
				fmt.Printf("error retrieving recipients by criteria: %s", err)
			}

			if len(criteriaRecipients) > 0 {
				if len(messageRecipients) > 0 {
					messageRecipients = getCommonRecipients(messageRecipients, criteriaRecipients)
				} else {
					messageRecipients = append(messageRecipients, criteriaRecipients...)
				}
			} else {
				messageRecipients = nil
			}
			log.Printf("construct message criteria recipients (%+v) for message (%s:%s:%s)", messageRecipients, *message.ID, message.Subject, message.Body)
		}

		if len(messageRecipients) > 0 {
			message.Recipients = messageRecipients
			if storeInInbox {
				persistedMessage, err = app.storage.UpdateMessage(message) // just update the message
				if err != nil {
					fmt.Printf("error storing the message: %s", err)
				} else {
					log.Printf("message %s has been updated", *persistedMessage.ID)
				}
			}

			// retrieve tokens by recipients
			tokens, err := app.storage.GetFirebaseTokensByRecipients(message.Recipients, message.RecipientsCriteriaList)
			if err != nil {
				log.Printf("error on GetFirebaseTokensByRecipients: %s", err)
				return nil, err
			}
			log.Printf("retrieve firebase tokens for message %s: %+v", *persistedMessage.ID, tokens)

			// send message to tokens
			if len(tokens) > 0 {
				if async {
					go app.sendNotifications(message, tokens)
				} else {
					app.sendNotifications(message, tokens)
				}
			}
		}

		if err != nil {
			fmt.Printf("create message finished with error: %s", err)
			return nil, err
		}

		return persistedMessage, err */
}

func (app *Application) sendNotifications(message *model.Message, tokens []string) {
	/*for _, token := range tokens {
		sendErr := app.firebase.SendNotificationToToken(token, message.Subject, message.Body, message.Data)
		if sendErr != nil {
			fmt.Printf("error send notification to token (%s): %s", token, sendErr)
		} else {
			log.Printf("message(%s:%s:%s) has been sent to token: %s", *message.ID, message.Subject, message.Body, token)
		}
	} */
}

func (app *Application) getMessages(userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	//return app.storage.GetMessages(userID, messageIDs, startDateEpoch, endDateEpoch, filterTopic, offset, limit, order)
	return nil, nil
}

func (app *Application) getMessage(ID string) (*model.Message, error) {
	return nil, nil
	//return app.storage.GetMessage(ID)
}

func (app *Application) updateMessage(user *model.CoreToken, message *model.Message) (*model.Message, error) {
	/*if message.ID != nil {
		persistedMessage, err := app.storage.GetMessage(*message.ID)
		if err == nil && persistedMessage != nil {
			if persistedMessage.Sender.User != nil && persistedMessage.Sender.User.UserID == user.UserID {
				return app.storage.UpdateMessage(message)
			}
			return nil, fmt.Errorf("only creator can update the original message")
		}
	}
	return nil, fmt.Errorf("missing id or record") */
	return nil, nil
}

func (app *Application) deleteUserMessage(user *model.CoreToken, messageID string) error {
	return nil
	//return app.storage.DeleteUserMessageWithContext(context.Background(), *user.UserID, messageID)
}

func (app *Application) deleteMessage(ID string) error {
	return nil
	//return app.storage.DeleteMessageWithContext(context.Background(), ID)
}

func (app *Application) getAllAppVersions() ([]model.AppVersion, error) {
	//return app.storage.GetAllAppVersions()
	return nil, nil
}

func (app *Application) getAllAppPlatforms() ([]model.AppPlatform, error) {
	//return app.storage.GetAllAppPlatforms()
	return nil, nil
}

func (app *Application) findUserByID(userID string) (*model.User, error) {
	//return app.storage.FindUserByID(userID)
	return nil, nil
}

func (app *Application) updateUserByID(userID string, notificationsDisabled bool) (*model.User, error) {
	//return app.storage.UpdateUserByID(userID, notificationsDisabled)
	return nil, nil
}

func (app *Application) deleteUserWithID(userID string) error {
	/*user, err := app.storage.FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("unable to delete user(%s): %s", userID, err)
	}

	if user != nil {
		err = app.storage.DeleteUserWithID(userID)
		if err != nil {
			return fmt.Errorf("unable to delete user(%s): %s", userID, err)
		}

		if user.Topics != nil && len(user.Topics) > 0 {
			for _, topic := range user.Topics {
				if user.FirebaseTokens != nil && len(user.FirebaseTokens) > 0 {
					for _, token := range user.FirebaseTokens {
						err := app.firebase.UnsubscribeToTopic(token.Token, topic)
						if err != nil {
							return fmt.Errorf("error unsubscribe user(%s) with token(%s) from topic(%s): %s", userID, token.Token, topic, err)
						}
					}
				}
			}
		}
	} */

	return nil
}

func getCommonRecipients(s1, s2 []model.Recipient) []model.Recipient {
	common := []model.Recipient{}
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[*e.UserID] = true
	}
	for _, e := range s2 {
		if hash[*e.UserID] == true {
			common = append(common, e)
		}
	}
	return common
}

func (app *Application) sendMail(toEmail string, subject string, body string) error {
	return app.mailer.SendMail(toEmail, subject, body)
}
