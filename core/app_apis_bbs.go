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
	"errors"
	"fmt"
	"notifications/core/model"
	"notifications/driven/storage"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/logs"
)

func (app *Application) bbsCreateMessages(inputMessages []model.InputMessage) ([]model.Message, error) {

	return app.sharedCreateMessages(inputMessages)
}

func (app *Application) bbsDeleteMessages(l *logs.Log, serviceAccountID string, messagesIDs []string) error {
	//in transaction
	transaction := func(context storage.TransactionContext) error {
		//find the messages
		messages, err := app.storage.FindMessagesWithContext(context, messagesIDs)
		if err != nil {
			return err
		}
		if len(messagesIDs) != len(messages) {
			return errors.New("not found message's")
		}
		/*
			//validate if the service account is the sender of this message
			valid := app.isSenderValid(serviceAccountID, *message)
			if !valid {
				return errors.New("not valid service account id for message - " + messageID)
			}*/

		//validate if the service account is the sender of the messages
		for _, m := range messages {
			valid := app.isSenderValid(serviceAccountID, m)
			if !valid {
				return errors.New("not valid service account id for message - " + m.ID)
			}
		}

		//delete the message
		messagesIDs := make([]string, len(messages))
		for i, m := range messages {
			messagesIDs[i] = m.ID
		}
		err = app.storage.DeleteMessagesWithContext(context, messagesIDs)
		if err != nil {
			return err
		}

		//delete the messages recipients
		err = app.storage.DeleteMessagesRecipientsForMessagesWithContext(context, messagesIDs)
		if err != nil {
			return err
		}

		//delete the queue data items
		err = app.storage.DeleteQueueDataForMessagesWithContext(context, messagesIDs)
		if err != nil {
			return err
		}

		return nil
	}

	//perform transactions
	err := app.storage.PerformTransaction(transaction, 2000)
	if err != nil {
		l.Errorf("error on performing delete message transaction - %s", err)
		return err
	}

	return nil
}

func (app *Application) isSenderValid(serviceAccountID string, message model.Message) bool {
	senderAccount := message.Sender.User
	if senderAccount == nil {
		return false
	}
	senderAccountID := senderAccount.UserID
	if senderAccountID != senderAccountID {
		return false
	}

	return true
}

func (app *Application) bbsSendMail(toEmail string, subject string, body string) error {
	return app.sharedSendMail(toEmail, subject, body)
}

func (app *Application) bbsAddRecipients(l *logs.Log, messageID string, orgID string, appID string, userID string, mute *bool, read *bool) ([]model.MessageRecipient, error) {
	var createRecipient []model.MessageRecipient
	notifyQueue := false
	//in transaction
	transaction := func(context storage.TransactionContext) error {
		//find the message
		message, err := app.storage.GetMessage(orgID, appID, messageID)
		if err != nil {
			return err
		}
		now := time.Now()
		//	now = "2023-03-21T08:01:15.294+00:00"
		recipientid := uuid.NewString()
		var recipient []model.MessageRecipient
		rec := model.MessageRecipient{OrgID: orgID, AppID: appID, ID: recipientid, UserID: userID,
			MessageID: messageID, Mute: *mute, Read: *read, Message: *message, DateCreated: &now}
		recipient = append(recipient, rec)
		createRecipient, err = app.storage.InsertMessagesRecipients(recipient)
		if err != nil {
			return err
		}

		//create the notifications queue items and store them in the queue
		queueItems := app.sharedCreateRecipientsQueueItems(message, recipient)
		if len(queueItems) > 0 {
			err = app.storage.InsertQueueDataRecipientsItems(queueItems)
			if err != nil {
				fmt.Printf("error on inserting queue data items: %s", err)
				return err
			}

			//notify the queue that new items are added
			notifyQueue = true
			return err
		}

		return nil
	}
	//perform transactions
	err := app.storage.PerformTransaction(transaction, 2000)
	if err != nil {
		l.Errorf("error on performing add message recipient transaction - %s", err)
		return nil, err
	}
	//notify the queue that new items are added
	if notifyQueue {
		go app.queueLogic.onQueuePush()
	}
	return createRecipient, nil
}

func (app *Application) bbsDeleteRecipients(l *logs.Log, id string, appID string, orgID string) error {
	//in transaction
	transaction := func(context storage.TransactionContext) error {

		err := app.storage.DeleteMessagesRecipients(id, appID, orgID)
		if err != nil {
			return err
		}

		//delete the queue data items
		err = app.storage.DeleteQueueDataForMessageRecipeint(id)
		if err != nil {
			return err
		}

		return nil
	}

	//perform transactions
	err := app.storage.PerformTransaction(transaction, 2000)
	if err != nil {
		l.Errorf("error on performing delete message recipient transaction - %s", err)
		return err
	}

	return nil
}
