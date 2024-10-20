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

func (app *Application) bbsCreateMessages(inputMessages []model.InputMessage, isBatch bool) ([]model.Message, error) {
	return app.sharedCreateMessages(inputMessages, isBatch)
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
	return senderAccountID == serviceAccountID
}

func (app *Application) bbsSendMail(toEmail string, subject string, body string) error {
	return app.sharedSendMail(toEmail, subject, body)
}

func (app *Application) bbsAddRecipients(l *logs.Log, serviceAccountID string, messageID string, inputRecipients []model.InputMessageRecipient) ([]model.MessageRecipient, error) {
	var err error
	var recipientsResult []model.MessageRecipient
	notifyQueue := false
	//in transaction
	transaction := func(context storage.TransactionContext) error {
		//find the message
		messageses, err := app.storage.FindMessagesWithContext(context, []string{messageID})
		if err != nil {
			return err
		}
		if len(messageses) == 0 {
			return errors.New("not found message")
		}
		message := messageses[0]

		//validate if the service account is the sender of the messages
		valid := app.isSenderValid(serviceAccountID, message)
		if !valid {
			return errors.New("not valid service account id for message - " + message.ID)
		}

		//create recipients objects
		recipients := make([]model.MessageRecipient, len(inputRecipients))
		for i, item := range inputRecipients {
			now := time.Now()
			current := model.MessageRecipient{OrgID: message.OrgID, AppID: message.AppID,
				ID: uuid.NewString(), UserID: item.UserID, MessageID: message.ID, Mute: item.Mute,
				Read: false, Message: message, DateCreated: &now}
			recipients[i] = current
		}

		//insert recipients
		err = app.storage.InsertMessagesRecipientsWithContext(context, recipients)
		if err != nil {
			fmt.Printf("error on inserting a recipient: %s", err)
			return err
		}

		//create the notifications queue items and store them in the queue
		queueItems := app.sharedCreateRecipientsQueueItems(&message, recipients)
		if len(queueItems) > 0 {
			err = app.storage.InsertQueueDataItemsWithContext(context, queueItems)
			if err != nil {
				fmt.Printf("error on inserting queue data items: %s", err)
				return err
			}
			//notify the queue that new items are added
			notifyQueue = true
			return err
		}

		recipientsResult = recipients

		return nil
	}
	//perform transactions
	err = app.storage.PerformTransaction(transaction, 10000) //10 seconds timeout
	if err != nil {
		fmt.Printf("error performing create recipients transaction - %s", err)
		return nil, err
	}

	//notify the queue that new items are added
	if notifyQueue {
		go app.queueLogic.onQueuePush()
	}

	return recipientsResult, nil
}

func (app *Application) bbsDeleteRecipients(l *logs.Log, serviceAccountID string, messageID string, usersIDs []string) error {
	//in transaction
	transaction := func(context storage.TransactionContext) error {
		//find the message
		messageses, err := app.storage.FindMessagesWithContext(context, []string{messageID})
		if err != nil {
			return err
		}
		if len(messageses) == 0 {
			return errors.New("not found message")
		}
		message := messageses[0]

		//validate if the service account is the sender of the messages
		valid := app.isSenderValid(serviceAccountID, message)
		if !valid {
			return errors.New("not valid service account id for message - " + message.ID)
		}

		//find the message recipients for deletion
		recipients, err := app.storage.FindMessagesRecipientsByMessageAndUsers(message.ID, usersIDs)
		if err != nil {
			return err
		}
		if len(recipients) != len(usersIDs) {
			return errors.New("not found recipient/s")
		}

		//prepare the messages recipients ids
		messagesRecipeintsIDs := make([]string, len(recipients))
		for i, item := range recipients {
			messagesRecipeintsIDs[i] = item.ID
		}

		//delete the messages recipients
		err = app.storage.DeleteMessagesRecipientsForIDsWithContext(context, messagesRecipeintsIDs)
		if err != nil {
			return err
		}

		//delete the queue data items
		err = app.storage.DeleteQueueDataForRecipientsWithContext(context, messagesRecipeintsIDs)
		if err != nil {
			return err
		}

		//notify the queue
		go app.queueLogic.onQueuePush()

		return nil
	}

	//perform transactions
	err := app.storage.PerformTransaction(transaction, 3000)
	if err != nil {
		l.Errorf("error on performing delete message recipient transaction - %s", err)
		return err
	}

	return nil
}
