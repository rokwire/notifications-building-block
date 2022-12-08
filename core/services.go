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
	"context"
	"fmt"
	"log"
	"notifications/core/model"
	"notifications/driven/storage"
	"time"

	"github.com/google/uuid"
)

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) storeFirebaseToken(orgID string, appID string, tokenInfo *model.TokenInfo, userID string) error {
	return app.storage.StoreFirebaseToken(orgID, appID, tokenInfo, userID)
}

func (app *Application) subscribeToTopic(orgID string, appID string, token string, userID string, anonymous bool, topic string) error {
	var err error
	if !anonymous {
		err = app.storage.SubscribeToTopic(orgID, appID, token, userID, topic)
		if err == nil {
			err = app.firebase.SubscribeToTopic(orgID, appID, token, topic)
		}
	} else if token != "" {
		// Treat this user as anonymous.
		err = app.firebase.SubscribeToTopic(orgID, appID, token, topic)
	}
	return err
}

func (app *Application) unsubscribeToTopic(orgID string, appID string, token string, userID string, anonymous bool, topic string) error {
	var err error
	if !anonymous {
		err = app.storage.UnsubscribeToTopic(orgID, appID, token, userID, topic)
		if err == nil {
			err = app.firebase.UnsubscribeToTopic(orgID, appID, token, topic)
		}
	} else if token != "" {
		// Treat this user as anonymous.
		err = app.firebase.UnsubscribeToTopic(orgID, appID, token, topic)
	}
	return err
}

func (app *Application) getTopics(orgID string, appID string) ([]model.Topic, error) {
	return app.storage.GetTopics(orgID, appID)
}

func (app *Application) appendTopic(topic *model.Topic) (*model.Topic, error) {
	return app.storage.InsertTopic(topic)
}

func (app *Application) updateTopic(topic *model.Topic) (*model.Topic, error) {
	return app.storage.UpdateTopic(topic)
}

func (app *Application) createMessage(inputMessage model.InputMessage, sender model.Sender, async bool) (*model.Message, error) {
	var err error
	var persistedMessage *model.Message
	var recipients []model.MessageRecipient

	//in transaction
	transaction := func(context storage.TransactionContext) error {

		//generate message id
		messageID := uuid.NewString()

		//calculate the recipients
		recipients, err = app.calculateRecipients(context, inputMessage, messageID)
		if err != nil {
			fmt.Printf("error on calculating recipients for a message: %s", err)
			return err
		}

		//create message object
		orgID := inputMessage.OrgID
		appID := inputMessage.AppID

		priority := inputMessage.Priority
		subject := inputMessage.Subject
		body := inputMessage.Body
		data := inputMessage.Data
		calculatedRecipients := len(recipients)
		var recipientsCriteriaList []model.RecipientCriteria
		if len(inputMessage.RecipientsCriteriaList) > 0 {
			recipientsCriteriaList = make([]model.RecipientCriteria, len(inputMessage.RecipientsCriteriaList))
			for i, item := range inputMessage.RecipientsCriteriaList {
				recipientsCriteriaList[i] = model.RecipientCriteria{AppVersion: item.AppVersion, AppPlatform: item.AppPlatform}
			}
		}
		topic := inputMessage.Topic
		dateCreated := time.Now()
		message := model.Message{OrgID: orgID, AppID: appID, ID: messageID, Priority: priority,
			Subject: subject, Sender: sender, Body: body, Data: data, RecipientsCriteriaList: recipientsCriteriaList,
			Topic: topic, CalculatedRecipientsCount: &calculatedRecipients, DateCreated: &dateCreated}

		//store the message object
		persistedMessage, err = app.storage.CreateMessageWithContext(context, message)
		if err != nil {
			fmt.Printf("error on creating a message: %s", err)
			return err
		}
		log.Printf("message %s has been created", persistedMessage.ID)

		//store recipients
		err = app.storage.InsertMessagesRecipientsWithContext(context, recipients)
		if err != nil {
			fmt.Printf("error on inserting recipients: %s", err)
			return err
		}

		return nil
	}

	//perform transactions
	err = app.storage.PerformTransaction(transaction, 10000) //10 seconds timeout
	if err != nil {
		fmt.Printf("error performing sync data transaction - %s", err)
		return nil, err
	}

	//send message
	if len(recipients) > 0 {
		err = app.sendMessage(recipients, *persistedMessage, async)
		if err != nil {
			fmt.Printf("error on sending message: %s", err)
			return nil, fmt.Errorf("error on sending message: %s", err)
		}
	}

	return persistedMessage, nil
}

func (app *Application) sendMessage(allRecipients []model.MessageRecipient, message model.Message, async bool) error {
	if len(allRecipients) == 0 {
		fmt.Print("no recipients")
		return nil
	}

	//send notifications only for mute=false
	recipients := []model.MessageRecipient{}
	for _, item := range allRecipients {
		if item.Mute == false {
			recipients = append(recipients, item)
		}
	}

	// retrieve tokens by recipients
	tokens, err := app.storage.GetFirebaseTokensByRecipients(
		message.OrgID, message.AppID, recipients, message.RecipientsCriteriaList)
	if err != nil {
		log.Printf("error on GetFirebaseTokensByRecipients: %s", err)
		return err
	}
	log.Printf("retrieve firebase tokens for message %s: %+v", message.ID, tokens)

	// send message to tokens
	if len(tokens) > 0 {
		if async {
			go app.sendNotifications(message, tokens)
		} else {
			app.sendNotifications(message, tokens)
		}
	}
	return nil
}

func (app *Application) calculateRecipients(context storage.TransactionContext,
	inputMessage model.InputMessage, messageID string) ([]model.MessageRecipient, error) {

	messageRecipients := []model.MessageRecipient{}
	checkCriteria := true

	// recipients from message
	if len(inputMessage.Recipients) > 0 {
		list := make([]model.MessageRecipient, len(inputMessage.Recipients))
		for i, item := range inputMessage.Recipients {
			cItem := model.MessageRecipient{OrgID: inputMessage.OrgID, AppID: inputMessage.AppID,
				ID: uuid.NewString(), UserID: item.UserID,
				MessageID: messageID, Mute: item.Mute, Read: false}
			list[i] = cItem
		}

		messageRecipients = append(messageRecipients, list...)
	}

	// recipients from topic
	if inputMessage.Topic != nil {
		topicUsers, err := app.storage.GetUsersByTopicWithContext(context, inputMessage.OrgID,
			inputMessage.AppID, *inputMessage.Topic)
		if err != nil {
			fmt.Printf("error retrieving recipients by topic (%s): %s", *inputMessage.Topic, err)
			return nil, err
		}
		log.Printf("retrieve recipients (%+v) for topic (%s)", topicUsers, *inputMessage.Topic)

		topicRecipients := make([]model.MessageRecipient, len(topicUsers))
		for i, item := range topicUsers {
			topicRecipients[i] = model.MessageRecipient{
				OrgID: inputMessage.OrgID, AppID: inputMessage.AppID,
				ID: uuid.NewString(), UserID: item.UserID, MessageID: messageID,
			}
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

		log.Printf("construct recipients (%+v) for message (%s:%s:%s)",
			messageRecipients, messageID, inputMessage.Subject, inputMessage.Body)
	}

	// recipients from criteria
	if (inputMessage.RecipientsCriteriaList != nil) && checkCriteria {
		criteriaList := make([]model.RecipientCriteria, len(inputMessage.RecipientsCriteriaList))
		for i, item := range inputMessage.RecipientsCriteriaList {
			criteriaList[i] = model.RecipientCriteria{AppVersion: item.AppVersion, AppPlatform: item.AppPlatform}
		}

		criteriaUsers, err := app.storage.GetUsersByRecipientCriteriasWithContext(context,
			inputMessage.OrgID, inputMessage.AppID, criteriaList)
		if err != nil {
			fmt.Printf("error retrieving recipients by criteria: %s", err)
			return nil, err
		}

		criteriaRecipients := make([]model.MessageRecipient, len(criteriaUsers))
		for i, item := range criteriaUsers {
			criteriaRecipients[i] = model.MessageRecipient{
				OrgID: inputMessage.OrgID, AppID: inputMessage.AppID,
				ID: uuid.NewString(), UserID: item.UserID, MessageID: messageID,
			}
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
		log.Printf("construct message criteria recipients (%+v) for message (%s:%s:%s)",
			messageRecipients, messageID, inputMessage.Subject, inputMessage.Body)
	}

	// recipients from account criteria
	if len(inputMessage.RecipientAccountCriteria) > 0 {
		accounts, err := app.core.RetrieveCoreUserAccountByCriteria(inputMessage.RecipientAccountCriteria,
			&inputMessage.AppID, &inputMessage.OrgID)
		if err != nil {
			fmt.Printf("error retrieving recipients by account criteria: %s", err)
		}

		for _, account := range accounts {
			messageRecipient := model.MessageRecipient{
				OrgID: inputMessage.OrgID, AppID: inputMessage.AppID,
				ID: uuid.NewString(), UserID: account.ID, MessageID: messageID,
			}

			messageRecipients = append(messageRecipients, messageRecipient)
		}

	}

	return messageRecipients, nil
}

func getCommonRecipients(s1, s2 []model.MessageRecipient) []model.MessageRecipient {
	common := []model.MessageRecipient{}
	messageReciepientsMap := map[string]model.MessageRecipient{}
	for _, e := range s1 {
		messageReciepientsMap[e.UserID] = e
	}
	for _, e := range s2 {
		if val, ok := messageReciepientsMap[e.UserID]; ok {
			common = append(common, val)
		}
	}
	return common
}

func (app *Application) sendNotifications(message model.Message, tokens []string) {
	for _, token := range tokens {
		sendErr := app.firebase.SendNotificationToToken(message.OrgID, message.AppID, token, message.Subject, message.Body, message.Data)
		if sendErr != nil {
			fmt.Printf("error send notification to token (%s): %s", token, sendErr)
		} else {
			log.Printf("message(%s:%s:%s) has been sent to token: %s", message.ID, message.Subject, message.Body, token)
		}
	}
}

func (app *Application) getMessages(orgID string, appID string, userID *string, read *bool, mute *bool, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	return app.storage.GetMessages(orgID, appID, userID, read, mute, messageIDs, startDateEpoch, endDateEpoch, filterTopic, offset, limit, order)
}

func (app *Application) getMessagesStats(orgID string, appID string, userID string) (*model.MessagesStats, error) {
	stats, _ := app.storage.GetMessagesStats(userID)
	return stats, nil
}

func (app *Application) getMessage(orgID string, appID string, ID string) (*model.Message, error) {
	return app.storage.GetMessage(orgID, appID, ID)
}

func (app *Application) getUserMessage(orgID string, appID string, ID string, accountID string) (*model.Message, error) {
	message, err := app.storage.GetMessage(orgID, appID, ID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		//no message for this id
		return nil, nil
	}

	//check if sender
	if message.IsSender(accountID) {
		return message, nil //it is sender
	}

	//check if recipient
	messagesRecipients, err := app.storage.FindMessagesRecipients(orgID, appID, ID, accountID)
	if err != nil {
		return nil, err
	}
	if len(messagesRecipients) > 0 {
		return message, err //it is recipient
	}

	return nil, nil //not sender, not recipient
}

func (app *Application) updateMessage(userID *string, message *model.Message) (*model.Message, error) {
	if message != nil {
		persistedMessage, err := app.storage.GetMessage(message.OrgID, message.AppID, message.ID)
		if err == nil && persistedMessage != nil {
			// If userID is nil, treat as system update, otherwise check sender match
			if userID == nil || (persistedMessage.Sender.User != nil && persistedMessage.Sender.User.UserID == *userID) {
				return app.storage.UpdateMessage(message)
			}
			return nil, fmt.Errorf("only creator can update the original message")
		}
	}
	return nil, fmt.Errorf("missing id or record")
}

func (app *Application) updateReadMessage(orgID string, appID string, ID string, userID string) (*model.Message, error) {
	updateReadMessage, _ := app.storage.UpdateUnreadMessage(context.Background(), orgID, appID, ID, userID)
	if updateReadMessage == nil {
		return nil, nil
	}
	return updateReadMessage, nil
}

func (app *Application) updateAllUserMessagesRead(orgID string, appID string, userID string, read bool) error {
	return app.storage.UpdateAllUserMessagesRead(context.Background(), orgID, appID, userID, read)
}

func (app *Application) deleteUserMessage(orgID string, appID string, userID string, messageID string) error {
	return app.storage.DeleteUserMessageWithContext(context.Background(), orgID, appID, userID, messageID)
}

func (app *Application) deleteMessage(orgID string, appID string, ID string) error {
	return app.storage.DeleteMessageWithContext(context.Background(), orgID, appID, ID)
}

func (app *Application) getAllAppVersions(orgID string, appID string) ([]model.AppVersion, error) {
	return app.storage.GetAllAppVersions(orgID, appID)
}

func (app *Application) getAllAppPlatforms(orgID string, appID string) ([]model.AppPlatform, error) {
	return app.storage.GetAllAppPlatforms(orgID, appID)
}

func (app *Application) findUserByID(orgID string, appID string, userID string) (*model.User, error) {
	return app.storage.FindUserByID(orgID, appID, userID)
}

func (app *Application) updateUserByID(orgID string, appID string, userID string, notificationsDisabled bool) (*model.User, error) {
	return app.storage.UpdateUserByID(orgID, appID, userID, notificationsDisabled)
}

func (app *Application) deleteUserWithID(orgID string, appID string, userID string) error {
	user, err := app.storage.FindUserByID(orgID, appID, userID)
	if err != nil {
		return fmt.Errorf("unable to delete user(%s): %s", userID, err)
	}

	if user != nil {
		err = app.storage.DeleteUserWithID(orgID, appID, userID)
		if err != nil {
			return fmt.Errorf("unable to delete user(%s): %s", userID, err)
		}

		if user.Topics != nil && len(user.Topics) > 0 {
			for _, topic := range user.Topics {
				if user.FirebaseTokens != nil && len(user.FirebaseTokens) > 0 {
					for _, token := range user.FirebaseTokens {
						err := app.firebase.UnsubscribeToTopic(orgID, appID, token.Token, topic)
						if err != nil {
							return fmt.Errorf("error unsubscribe user(%s) with token(%s) from topic(%s): %s", userID, token.Token, topic, err)
						}
					}
				}
			}
		}
	}

	return nil
}

func (app *Application) sendMail(toEmail string, subject string, body string) error {
	return app.mailer.SendMail(toEmail, subject, body)
}
