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

	"github.com/rokwire/logging-library-go/v2/logs"
)

type queueLogic struct {
	logger *logs.Logger

	storage  Storage
	firebase Firebase
}

func (q queueLogic) start() {
	q.logger.Info("queueLogic start")

	q.processQueue()
}

func (q queueLogic) onQueuePush() {
	q.logger.Info("queueLogic onQueuePush")

	q.processQueue()
}

func (q queueLogic) processQueue() {
	q.logger.Info("queueLogic processQueue")

	// check if the queue is locked and lock it for processing
	queueAvailable, queue, err := q.lockQueue()
	if err != nil {
		q.logger.Errorf("error on locking queue", err)
		return
	}
	if !*queueAvailable {
		q.logger.Info("the queue is locked, so do nothing")
		return
	}

	//TODO

	time := time.Now()
	limit := queue.ProcessItemsCount
	queueItems, err := q.storage.FindQueueData(time, limit)
	if err != nil {
		q.logger.Errorf("error on finding queue data", err)

		q.unlockQueue(*queue) //always unlock the queue on error
		return
	}

	log.Println(queueItems)

	//TODO set timer
}

func (q queueLogic) lockQueue() (*bool, *model.Queue, error) {
	var err error
	var queue *model.Queue
	queueAvailable := true

	// in transaction
	transaction := func(context storage.TransactionContext) error {
		//load queue
		queue, err = q.storage.LoadQueueWithContext(context)
		if err != nil {
			q.logger.Infof("error on loading queue: %s", err)
			return err
		}

		//check if available
		if queue.Status != "ready" {
			q.logger.Infof("the queue is not ready but %s", queue.Status)
			queueAvailable = false
			return nil
		}

		//lock it
		queue.Status = "processing"
		err = q.storage.SaveQueueWithContext(context, *queue)
		if err != nil {
			q.logger.Infof("error on marking the queue locked: %s", err)
			return err
		}

		return nil
	}
	//perform transactions
	err = q.storage.PerformTransaction(transaction, 2000)
	if err != nil {
		fmt.Printf("error performing lock queue transaction - %s", err)
		return nil, nil, err
	}

	return &queueAvailable, queue, nil
}

func (q queueLogic) unlockQueue(queue model.Queue) {
	queue.Status = "ready"
	err := q.storage.SaveQueue(queue)
	if err != nil {
		q.logger.Errorf("error unlocking the queue - %s", err) //cannot be done anything else
	}
}

/* TODO
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
} */

/*
func (app *Application) sendNotifications(message model.Message, tokens []string) {
	for _, token := range tokens {
		sendErr := app.firebase.SendNotificationToToken(message.OrgID, message.AppID, token, message.Subject, message.Body, message.Data)
		if sendErr != nil {
			fmt.Printf("error send notification to token (%s): %s", token, sendErr)
		} else {
			log.Printf("message(%s:%s:%s) has been sent to token: %s", message.ID, message.Subject, message.Body, token)
		}
	}
} */
