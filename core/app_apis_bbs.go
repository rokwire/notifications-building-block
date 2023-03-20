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
	"notifications/core/model"
	"notifications/driven/storage"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/logs"
)

func (app *Application) bbsCreateMessage(inputMessage model.InputMessage) (*model.Message, error) {

	return app.sharedCreateMessage(inputMessage)
}

func (app *Application) bbsDeleteMessage(l *logs.Log, serviceAccountID string, messageID string) error {
	//in transaction
	transaction := func(context storage.TransactionContext) error {
		//find the message
		message, err := app.storage.FindMessageWithContext(context, messageID)
		if err != nil {
			return err
		}
		if message == nil {
			return errors.New("no message for id - " + messageID)
		}

		//validate if the service account is the sender of this message
		valid := app.isSenderValid(serviceAccountID, *message)
		if !valid {
			return errors.New("not valid service account id for message - " + messageID)
		}

		//delete the message
		err = app.storage.DeleteMessageWithContext(context, message.OrgID, message.AppID, messageID)
		if err != nil {
			return err
		}

		//delete the message recipients
		err = app.storage.DeleteMessagesRecipientsForMessageWithContext(context, messageID)
		if err != nil {
			return err
		}

		//delete the queue data items
		err = app.storage.DeleteQueueDataForMessageWithContext(context, messageID)
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
	//in transaction
	transaction := func(context storage.TransactionContext) error {
		//find the message
		message, err := app.storage.GetMessage(orgID, appID, messageID)
		if err != nil {
			return err
		}
		now := time.Now()
		recipientid := uuid.NewString()
		var recipient []model.MessageRecipient
		rec := model.MessageRecipient{OrgID: orgID, AppID: appID, ID: recipientid, UserID: userID,
			MessageID: messageID, Mute: *mute, Read: *read, Message: *message, DateCreated: &now}
		recipient = append(recipient, rec)
		createRecipient, err = app.storage.InsertMessagesRecipients(recipient)
		if err != nil {
			return err
		}
		return err
	}

	//perform transactions
	err := app.storage.PerformTransaction(transaction, 2000)
	if err != nil {
		l.Errorf("error on performing delete message transaction - %s", err)
		return nil, err
	}

	return createRecipient, nil
}

func (app *Application) bbsDeleteRecipients(l *logs.Log, orgID string, appID string, messageID string) error {
	//in transaction
	transaction := func(context storage.TransactionContext) error {
		//find the message
		message, err := app.storage.GetMessage(orgID, appID, messageID)
		if err != nil {
			return err
		}

		if message.Recipients != nil {
			err := app.storage.DeleteRecipientsFromMessage(message.Recipients, messageID)
			if err != nil {
				return err
			}
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
