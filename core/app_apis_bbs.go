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

	"github.com/rokwire/logging-library-go/v2/logs"
)

func (app *Application) bbsCreateMessage(orgID string, appID string,
	sender model.Sender, mTime time.Time, priority int, subject string, body string, data map[string]string,
	inputRecipients []model.MessageRecipient, recipientsCriteriaList []model.RecipientCriteria,
	recipientAccountCriteria map[string]interface{}, topic *string, async bool) (*model.Message, error) {

	return app.sharedCreateMessage(orgID, appID, sender, mTime, priority, subject, body, data,
		inputRecipients, recipientsCriteriaList, recipientAccountCriteria, topic, async)
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
		//TODO

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
