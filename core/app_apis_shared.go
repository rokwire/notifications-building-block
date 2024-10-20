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
	"fmt"
	"log"
	"notifications/core/model"
	"notifications/driven/storage"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
)

func (app *Application) sharedCreateMessages(imMessages []model.InputMessage, isBatch bool) ([]model.Message, error) {

	if len(imMessages) == 0 {
		return nil, errors.New("no data")
	}

	var err error
	resultMessages := []model.Message{}
	notifyQueue := false

	//in transaction
	transaction := func(context storage.TransactionContext) error {

		allMessages := []model.Message{}
		allRecipients := []model.MessageRecipient{}
		allQueueItems := []model.QueueItem{}

		recipientsMap := map[string]bool{}

		//process every message
		for _, im := range imMessages {
			message, recipients, err := app.sharedHandleInputMessage(context, im)
			if err != nil {
				fmt.Printf("error on handling a message: %s", err)
				return err
			}
			// if batched messages, only send single highest priority (first) message to each recipient
			if isBatch {
				batchRecipients := []model.MessageRecipient{}
				for _, recipient := range recipients {
					if _, ok := recipientsMap[recipient.UserID]; !ok {
						recipientsMap[recipient.UserID] = true
						batchRecipients = append(batchRecipients, recipient)
					}
				}
				recipients = batchRecipients
				recipientCount := len(recipients)
				message.CalculatedRecipientsCount = &recipientCount
			}
			queueItems := app.sharedCreateQueueItems(*message, recipients)
			allMessages = append(allMessages, *message)
			allRecipients = append(allRecipients, recipients...)
			allQueueItems = append(allQueueItems, queueItems...)
		}

		//store the messages object
		err = app.storage.InsertMessagesWithContext(context, allMessages)
		if err != nil {
			fmt.Printf("error on creating a message: %s", err)
			return err
		}

		//store recipients
		err = app.storage.InsertMessagesRecipientsWithContext(context, allRecipients)
		if err != nil {
			fmt.Printf("error on inserting recipients: %s", err)
			return err
		}

		//store the notifications queue items in the queue
		if len(allQueueItems) > 0 {
			err = app.storage.InsertQueueDataItemsWithContext(context, allQueueItems)
			if err != nil {
				fmt.Printf("error on inserting queue data items: %s", err)
				return err
			}

			//notify the queue that new items are added
			notifyQueue = true
		}

		resultMessages = allMessages

		return nil
	}

	//perform transactions
	err = app.storage.PerformTransaction(transaction, 10000) //10 seconds timeout
	if err != nil {
		fmt.Printf("error performing create message transaction - %s", err)
		return nil, err
	}

	//notify the queue that new items are added
	if notifyQueue {
		go app.queueLogic.onQueuePush()
	}

	return resultMessages, nil
}

func (app *Application) sharedHandleInputMessage(context storage.TransactionContext, im model.InputMessage) (*model.Message, []model.MessageRecipient, error) {
	//use from input if available
	messageID := im.ID
	if messageID == nil {
		genMessageID := uuid.NewString()
		messageID = &genMessageID
	}

	//calculate the recipients
	recipients, err := app.sharedCalculateRecipients(context, im.OrgID, im.AppID,
		im.Subject, im.Body, im.InputRecipients, im.RecipientsCriteriaList,
		im.RecipientAccountCriteria, im.Topics, *messageID)
	if err != nil {
		fmt.Printf("error on calculating recipients for a message: %s", err)
		return nil, nil, err
	}

	//create message object
	if im.Data == nil { //we add message id to the data
		im.Data = map[string]string{}
	}
	im.Data["message_id"] = *messageID
	calculatedRecipients := len(recipients)
	dateCreated := time.Now()
	message := model.Message{OrgID: im.OrgID, AppID: im.AppID, ID: *messageID, Priority: im.Priority, Time: im.Time,
		Subject: im.Subject, Sender: im.Sender, Body: im.Body, Data: im.Data, RecipientsCriteriaList: im.RecipientsCriteriaList,
		RecipientAccountCriteria: im.RecipientAccountCriteria, Topic: im.Topic, Topics: im.Topics,
		CalculatedRecipientsCount: &calculatedRecipients, DateCreated: &dateCreated}

	return &message, recipients, nil
}

func (app *Application) sharedCreateQueueItems(message model.Message, messageRecipients []model.MessageRecipient) []model.QueueItem {
	queueItems := []model.QueueItem{}

	for _, messageRecipient := range messageRecipients {
		orgID := messageRecipient.OrgID
		appID := messageRecipient.AppID
		id := uuid.NewString()

		messageID := message.ID

		messageRecipientID := messageRecipient.ID
		userID := messageRecipient.UserID

		subject := message.Subject
		body := message.Body
		data := message.Data

		time := message.Time
		priority := message.Priority

		queueItem := model.QueueItem{OrgID: orgID, AppID: appID, ID: id,
			MessageID: messageID, MessageRecipientID: messageRecipientID, UserID: userID,
			Subject: subject, Body: body, Data: data, Time: time, Priority: priority}

		queueItems = append(queueItems, queueItem)
	}

	return queueItems
}

func (app *Application) sharedCalculateRecipients(context storage.TransactionContext,
	orgID string, appID string,
	subject string, body string,
	recipients []model.MessageRecipient, recipientsCriteriaList []model.RecipientCriteria,
	recipientAccountCriteria map[string]interface{}, topics []string, messageID string) ([]model.MessageRecipient, error) {

	messageRecipients := []model.MessageRecipient{}
	checkCriteria := true
	now := time.Now()

	// recipients from message
	if len(recipients) > 0 {
		list := make([]model.MessageRecipient, len(recipients))
		for i, item := range recipients {
			item.OrgID = orgID
			item.AppID = appID
			item.ID = uuid.NewString()
			item.MessageID = messageID
			item.Read = false
			item.DateCreated = &now

			list[i] = item
		}

		messageRecipients = append(messageRecipients, list...)
	}

	// recipients from topic
	if topics != nil {
		topicUsers, err := app.storage.GetUsersByTopicsWithContext(context, orgID,
			appID, topics)
		if err != nil {
			fmt.Printf("error retrieving recipients by topic (%s): %s", topics, err)
			return nil, err
		}
		log.Printf("retrieve recipients (%+v) for topic (%s)", topicUsers, topics)

		topicRecipients := make([]model.MessageRecipient, len(topicUsers))
		for i, item := range topicUsers {
			topicRecipients[i] = model.MessageRecipient{
				OrgID: orgID, AppID: appID, ID: uuid.NewString(), UserID: item.UserID,
				MessageID: messageID, DateCreated: &now,
			}
		}

		if len(topicRecipients) > 0 {
			if len(messageRecipients) > 0 {
				messageRecipients = sharedGetCommonRecipients(messageRecipients, topicRecipients)
			} else {
				messageRecipients = append(messageRecipients, topicRecipients...)
			}
		} else {
			checkCriteria = false
			messageRecipients = nil
		}

		log.Printf("construct recipients (%+v) for message (%s:%s:%s)",
			messageRecipients, messageID, subject, body)
	}

	// recipients from criteria
	if len(recipientsCriteriaList) > 0 && checkCriteria {
		criteriaUsers, err := app.storage.GetUsersByRecipientCriteriasWithContext(context,
			orgID, appID, recipientsCriteriaList)
		if err != nil {
			fmt.Printf("error retrieving recipients by criteria: %s", err)
			return nil, err
		}

		criteriaRecipients := make([]model.MessageRecipient, len(criteriaUsers))
		for i, item := range criteriaUsers {
			criteriaRecipients[i] = model.MessageRecipient{
				OrgID: orgID, AppID: appID, ID: uuid.NewString(), UserID: item.UserID,
				MessageID: messageID, DateCreated: &now,
			}
		}

		if len(criteriaRecipients) > 0 {
			if len(messageRecipients) > 0 {
				messageRecipients = sharedGetCommonRecipients(messageRecipients, criteriaRecipients)
			} else {
				messageRecipients = append(messageRecipients, criteriaRecipients...)
			}
		} else {
			messageRecipients = nil
		}
		log.Printf("construct message criteria recipients (%+v) for message (%s:%s:%s)",
			messageRecipients, messageID, subject, body)
	}

	// recipients from account criteria
	if len(recipientAccountCriteria) > 0 {
		accounts, err := app.core.RetrieveCoreUserAccountByCriteria(recipientAccountCriteria,
			&appID, &orgID)
		if err != nil {
			fmt.Printf("error retrieving recipients by account criteria: %s", err)
		}

		for _, account := range accounts {
			messageRecipient := model.MessageRecipient{
				OrgID: orgID, AppID: appID, ID: uuid.NewString(), UserID: account.ID,
				MessageID: messageID, DateCreated: &now,
			}

			messageRecipients = append(messageRecipients, messageRecipient)
		}

	}

	return messageRecipients, nil
}

func sharedGetCommonRecipients(messageRecipients, topicRecipients []model.MessageRecipient) []model.MessageRecipient {
	//
	// Recipients who don't belong to a topic will still receive a muted message (just skipping the push notification)
	//
	common := []model.MessageRecipient{}
	topicReciepientsMap := map[string]model.MessageRecipient{}

	for _, e := range topicRecipients {
		topicReciepientsMap[e.UserID] = e
	}

	for _, recipient := range messageRecipients {
		if _, ok := topicReciepientsMap[recipient.UserID]; !ok {
			recipient.Mute = true
		}
		common = append(common, recipient)
	}

	return common
}

func (app *Application) sharedSendMail(toEmail string, subject string, body string) error {
	return app.mailer.SendMail(toEmail, subject, body)
}

func (app *Application) sharedCreateRecipientsQueueItems(message *model.Message, messageRecipients []model.MessageRecipient) []model.QueueItem {
	queueItems := []model.QueueItem{}

	for _, messageRecipient := range messageRecipients {
		orgID := messageRecipient.OrgID
		appID := messageRecipient.AppID
		id := messageRecipient.ID
		userID := messageRecipient.UserID
		messageID := messageRecipient.MessageID
		subject := message.Subject
		body := message.Body
		data := message.Data
		time := message.Time
		priority := message.Priority

		queueItem := model.QueueItem{OrgID: orgID, AppID: appID, ID: id,
			MessageID: messageID, MessageRecipientID: id, UserID: userID, Subject: subject, Body: body,
			Data: data, Time: time, Priority: priority}

		queueItems = append(queueItems, queueItem)
	}

	return queueItems
}
