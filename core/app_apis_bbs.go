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
