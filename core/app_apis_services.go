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
	"errors"
	"fmt"
	"notifications/core/model"
	"sync"
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

func (app *Application) getUserData(orgID, appID, userID string) (*model.UserDataResponse, error) {
	var (
		receivedNotifications       []model.Message
		scheduledNotificationsForMe []model.Message
		recipientData               []model.MessageRecipient
		queueData                   []model.QueueItem
		user                        *model.User
		err                         error
	)

	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, 3)

	// Fetch recipient data concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		recipientData, err = app.storage.FindMessagesRecipientsByUserID(orgID, appID, userID)
		if err != nil {
			errCh <- err
		}
	}()

	// Fetch queue data concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		queueData, err = app.storage.FindQueueDataByUserID(userID)
		if err != nil {
			errCh <- err
		}
	}()

	// Fetch user data concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		user, err = app.storage.FindUserByID(orgID, appID, userID)
		if err != nil {
			errCh <- err
		}
	}()

	wg.Wait()
	close(errCh)

	// Check for errors
	for e := range errCh {
		if e != nil {
			return nil, e
		}
	}

	// Fetch messages related to recipient data
	if recipientData != nil {
		for _, rn := range recipientData {
			wg.Add(1)
			go func(rn model.MessageRecipient) {
				defer wg.Done()
				rnr, err := app.storage.GetMessage(rn.OrgID, rn.AppID, rn.MessageID)
				if err == nil && rnr != nil {
					mu.Lock()
					receivedNotifications = append(receivedNotifications, *rnr)
					mu.Unlock()
				}
			}(rn)
		}
	}

	// Fetch messages related to queue data
	if queueData != nil {
		for _, q := range queueData {
			wg.Add(1)
			go func(q model.QueueItem) {
				defer wg.Done()
				qr, err := app.storage.GetMessage(q.OrgID, q.AppID, q.MessageID)
				if err == nil && qr != nil {
					mu.Lock()
					scheduledNotificationsForMe = append(scheduledNotificationsForMe, *qr)
					mu.Unlock()
				}
			}(q)
		}
	}

	wg.Wait()

	userData := &model.UserDataResponse{
		ReceivedNotifications:       receivedNotifications,
		ScheduledNotificationsForMe: scheduledNotificationsForMe,
		Users:                       *user,
	}
	return userData, nil
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

func (app *Application) createMessage(inputMessage model.InputMessage) (*model.Message, error) {
	inputMessages := []model.InputMessage{inputMessage} //only one
	messages, err := app.sharedCreateMessages(inputMessages)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, errors.New("error on creating message")
	}

	return &messages[0], nil //return only one
}

func (app *Application) getMessagesRecipientsDeep(orgID string, appID string, userID *string, read *bool, mute *bool, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.MessageRecipient, error) {
	return app.storage.FindMessagesRecipientsDeep(orgID, appID, userID, read, mute, messageIDs, startDateEpoch, endDateEpoch, filterTopic, offset, limit, order)
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
	return app.storage.DeleteMessagesWithContext(context.Background(), []string{ID})
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
	return app.sharedSendMail(toEmail, subject, body)
}
