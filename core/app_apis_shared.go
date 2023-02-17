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
)

func (app *Application) sharedCreateMessage(orgID string, appID string,
	sender model.Sender, mTime time.Time, priority int, subject string, body string, data map[string]string,
	inputRecipients []model.MessageRecipient, recipientsCriteriaList []model.RecipientCriteria,
	recipientAccountCriteria map[string]interface{}, topic *string, async bool) (*model.Message, error) {

	var err error
	var persistedMessage *model.Message
	var recipients []model.MessageRecipient
	notifyQueue := false

	//in transaction
	transaction := func(context storage.TransactionContext) error {

		//generate message id
		messageID := uuid.NewString()

		//calculate the recipients
		recipients, err = app.sharedCalculateRecipients(context, orgID, appID,
			subject, body, inputRecipients, recipientsCriteriaList,
			recipientAccountCriteria, topic, messageID)
		if err != nil {
			fmt.Printf("error on calculating recipients for a message: %s", err)
			return err
		}

		//create message object
		if data == nil { //we add message id to the data
			data = map[string]string{}
		}
		data["message_id"] = messageID
		calculatedRecipients := len(recipients)
		dateCreated := time.Now()
		message := model.Message{OrgID: orgID, AppID: appID, ID: messageID, Priority: priority, Time: mTime,
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

		//create the notifications queue items and store them in the queue
		queueItems := app.sharedCreateQueueItems(*persistedMessage, recipients)
		if len(queueItems) > 0 {
			err = app.storage.InsertQueueDataItemsWithContext(context, queueItems)
			if err != nil {
				fmt.Printf("error on inserting queue data items: %s", err)
				return err
			}

			//notify the queue that new items are added
			notifyQueue = true
		}

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

	return persistedMessage, nil
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
	recipientAccountCriteria map[string]interface{}, topic *string, messageID string) ([]model.MessageRecipient, error) {

	messageRecipients := []model.MessageRecipient{}
	checkCriteria := true

	// recipients from message
	if len(recipients) > 0 {
		list := make([]model.MessageRecipient, len(recipients))
		for i, item := range recipients {
			item.OrgID = orgID
			item.AppID = appID
			item.ID = uuid.NewString()
			item.MessageID = messageID
			item.Read = false

			list[i] = item
		}

		messageRecipients = append(messageRecipients, list...)
	}

	// recipients from topic
	if topic != nil {
		topicUsers, err := app.storage.GetUsersByTopicWithContext(context, orgID,
			appID, *topic)
		if err != nil {
			fmt.Printf("error retrieving recipients by topic (%s): %s", *topic, err)
			return nil, err
		}
		log.Printf("retrieve recipients (%+v) for topic (%s)", topicUsers, *topic)

		topicRecipients := make([]model.MessageRecipient, len(topicUsers))
		for i, item := range topicUsers {
			topicRecipients[i] = model.MessageRecipient{
				OrgID: orgID, AppID: appID,
				ID: uuid.NewString(), UserID: item.UserID, MessageID: messageID,
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
	if (recipientsCriteriaList != nil && len(recipientAccountCriteria) > 0) && checkCriteria {
		criteriaUsers, err := app.storage.GetUsersByRecipientCriteriasWithContext(context,
			orgID, appID, recipientsCriteriaList)
		if err != nil {
			fmt.Printf("error retrieving recipients by criteria: %s", err)
			return nil, err
		}

		criteriaRecipients := make([]model.MessageRecipient, len(criteriaUsers))
		for i, item := range criteriaUsers {
			criteriaRecipients[i] = model.MessageRecipient{
				OrgID: orgID, AppID: appID,
				ID: uuid.NewString(), UserID: item.UserID, MessageID: messageID,
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
				OrgID: orgID, AppID: appID,
				ID: uuid.NewString(), UserID: account.ID, MessageID: messageID,
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
