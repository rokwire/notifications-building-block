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
	"notifications/driven/storage"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

type queueLogic struct {
	logger *logs.Logger

	storage  Storage
	firebase Firebase

	//timer
	queueTimer *time.Timer
	timerDone  chan bool
}

func (q *queueLogic) start() {
	q.logger.Info("queueLogic start")

	q.processQueue()
}

func (q *queueLogic) onQueuePush() {
	q.logger.Info("queueLogic onQueuePush")

	q.processQueue()
}

func (q *queueLogic) processQueue() {
	q.logger.Info("queueLogic processQueue")

	//check if the queue is locked and lock it for processing
	queueAvailable, queue, err := q.lockQueue()
	if err != nil {
		q.logger.Errorf("error on locking queue - %s", err)
		return
	}
	if !*queueAvailable {
		q.logger.Info("the queue is locked, so do nothing")
		return
	}

	//ensure the queue is always unlocked and protect against panics
	defer func() {
		if r := recover(); r != nil {
			q.logger.Errorf("panic in processQueue: %v", r)
		}
		q.unlockQueue(*queue)
	}()

	//process the queue items until they are available
	now := time.Now()
	limit := queue.ProcessItemsCount
	for {
		//get the current items
		queueItems, err := q.storage.FindQueueData(&now, limit)
		if err != nil {
			q.logger.Errorf("error on finding queue data - %s", err)
			return
		}

		itemsCount := len(queueItems)
		if itemsCount == 0 {
			q.logger.Info("no more items for processing, stop iterating")
			break //no more items for processing, stop iterating
		}

		q.logger.Infof("%d items to processes", itemsCount)

		//process the current items
		err = q.processQueueItem(queueItems)
		if err != nil {
			q.logger.Errorf("error on processing items - %s", err)
			return
		}
	}

	//set timer if there is still items in the queue for scheduled messages
	err = q.setTimerIfNecessary()
	if err != nil {
		q.logger.Errorf("error on setting timer - %s", err)
		return
	}
}

func (q *queueLogic) setTimerIfNecessary() error {
	//check if there is scheduled messages
	scheduled, err := q.storage.FindQueueData(nil, 1) //it gives the first upcoming message
	if err != nil {
		return err
	}

	if len(scheduled) == 0 {
		q.logger.Info("there is no upcoming messages in the queue, so not setting timer")
		return nil
	}

	upcomingTime := scheduled[0].Time
	q.logger.Infof("there is upcoming message at - %s", upcomingTime)

	//set timer
	go q.setTimer(upcomingTime)

	return nil
}

func (q *queueLogic) setTimer(upcomingTime time.Time) error {
	nowInSeconds := time.Now().Unix()
	upcomingInSeconds := upcomingTime.Unix()
	durationInSeconds := (upcomingInSeconds - nowInSeconds) + 2 //add two seconds to be sure that the timer will be executed after the message time
	duration := time.Second * time.Duration(durationInSeconds)

	q.logger.Infof("setting timer after - %s", duration)

	q.queueTimer = time.NewTimer(duration)
	select {
	case <-q.queueTimer.C:
		q.logger.Info("setTimer -> queue timer expired")
		q.queueTimer = nil

		q.processQueue()
	case <-q.timerDone:
		// timer aborted
		q.logger.Info("setTimer -> queue timer aborted")
		q.queueTimer = nil
	}

	return nil
}

func (q *queueLogic) lockQueue() (*bool, *model.Queue, error) {
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
		q.logger.Errorf("error performing lock queue transaction - %s", err)
		return nil, nil, err
	}

	return &queueAvailable, queue, nil
}

func (q *queueLogic) unlockQueue(queue model.Queue) {
	queue.Status = "ready"
	var err error
	for i := 0; i < 3; i++ {
		err = q.storage.SaveQueue(queue)
		if err == nil {
			return
		}
		q.logger.Errorf("error unlocking the queue (attempt %d) - %s", i+1, err)
		time.Sleep(200 * time.Millisecond)
	}
	q.logger.Errorf("failed to unlock the queue after retries - %s", err)
}

func (q *queueLogic) processQueueItem(queueItems []model.QueueItem) error {

	//get the users as we need their tokens and if they have disabled notifications
	usersIDs := make([]string, len(queueItems))
	for i, item := range queueItems {
		usersIDs[i] = item.UserID
	}
	users, err := q.storage.FindUsersByIDs(usersIDs)
	if err != nil {
		q.logger.Errorf("error on getting users - %s", err)
		return err
	}

	//process every item
	itemsIDs := make([]string, len(queueItems))
	for i, item := range queueItems {
		itemsIDs[i] = item.ID

		var user *model.User

		//get the user
		for _, cUser := range users {
			if cUser.UserID == item.UserID {
				user = &cUser
				break
			}
		}

		if user == nil {
			continue //for some reasons there is no a corresponding user
		}

		if user.NotificationsDisabled {
			continue //do not send notification if disabled for the user
		}

		tokens := user.FirebaseTokens
		go q.sendNotifications(item, tokens) //new thread
	}

	//remove the items from the queue
	err = q.storage.DeleteQueueData(itemsIDs)
	if err != nil {
		q.logger.Errorf("error on deleting queue datas - %s", err)
		return err
	}

	return nil
}

func (q *queueLogic) sendNotifications(queueItem model.QueueItem, tokens []model.FirebaseToken) {

	for _, fToken := range tokens {
		token := fToken.Token
		sendErr := q.firebase.SendNotificationToToken(queueItem.OrgID, queueItem.AppID, token, queueItem.Subject, queueItem.Body, queueItem.Data)
		if sendErr != nil {
			q.logger.Errorf("error send notification to token (%s): %s", token, sendErr)
		} else {
			q.logger.Infof("queue item(%s:%s:%s) has been sent to token: %s", queueItem.ID, queueItem.Subject, queueItem.Body, token)
		}
	}
}
