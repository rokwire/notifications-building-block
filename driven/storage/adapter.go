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

package storage

import (
	"context"
	"fmt"
	"log"
	"notifications/core/model"
	"notifications/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"
	"golang.org/x/sync/syncmap"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Adapter implements the Storage interface
type Adapter struct {
	db *database

	cachedConfigs *syncmap.Map
	configsLock   *sync.RWMutex
}

// Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()

	//cache the configs
	err = sa.cacheConfigs()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCache, model.TypeConfig, nil, err)
	}

	return err
}

// RegisterStorageListener registers a data change listener with the storage adapter
func (sa *Adapter) RegisterStorageListener(storageListener Listener) {
	sa.db.listeners = append(sa.db.listeners, storageListener)
}

// PerformTransaction performs a transaction
func (sa *Adapter) PerformTransaction(transaction func(context TransactionContext) error, timeoutMilliSeconds int64) error {
	// transaction
	timeout := time.Millisecond * time.Duration(timeoutMilliSeconds)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := sa.db.dbClient.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction(logutils.ActionStart, logutils.TypeTransaction, nil, err)
		}

		err = transaction(sessionContext)
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction("performing", logutils.TypeTransaction, nil, err)
		}

		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction(logutils.ActionCommit, logutils.TypeTransaction, nil, err)
		}
		return nil
	})

	return err
}

// NewStorageAdapter creates a new storage adapter instance
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string,
	multiTenancyOrgID string, multiTenancyAppID string, logger *logs.Logger) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		log.Println("Set default timeout - 2000")
		timeout = 2000
	}
	timeoutMS := time.Millisecond * time.Duration(timeout)

	cachedConfigs := &syncmap.Map{}
	configsLock := &sync.RWMutex{}

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS,
		multiTenancyOrgID: multiTenancyOrgID, multiTenancyAppID: multiTenancyAppID, logger: logger}
	return &Adapter{db: db, cachedConfigs: cachedConfigs, configsLock: configsLock}
}

// LoadFirebaseConfigurations loads all firebase configurations
func (sa Adapter) LoadFirebaseConfigurations() ([]model.FirebaseConf, error) {
	filter := bson.D{}
	var result []model.FirebaseConf
	err := sa.db.firebaseConfigurations.Find(filter, &result, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "firebase configuration", nil, err)
	}
	return result, nil
}

// FindUsersByIDs finds users by ids
func (sa Adapter) FindUsersByIDs(usersIDs []string) ([]model.User, error) {
	filter := bson.D{
		primitive.E{Key: "user_id", Value: bson.M{"$in": usersIDs}},
	}

	var result []model.User
	err := sa.db.users.Find(filter, &result, nil)
	if err != nil {
		log.Printf("warning: error while retriving users - %s", err)
		return nil, err
	}

	return result, err
}

// FindUserByToken finds firebase token
func (sa Adapter) FindUserByToken(orgID string, appID string, token string) (*model.User, error) {
	return sa.findUserByTokenWithContext(context.Background(), orgID, appID, token)
}

func (sa Adapter) findUserByTokenWithContext(context context.Context, orgID string, appID string, token string) (*model.User, error) {
	filter := bson.D{}
	if len(token) > 0 {
		filter = bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "firebase_tokens", Value: bson.D{primitive.E{Key: "$elemMatch", Value: bson.D{primitive.E{Key: "token", Value: token}}}}},
		}
	}

	var result *model.User
	err := sa.db.users.FindOneWithContext(context, filter, &result, nil)
	if err != nil {
		log.Printf("warning: error while retrieving token (%s) - %s", token, err)
	}

	return result, err
}

// FindUserByID finds user by id
func (sa Adapter) FindUserByID(orgID string, appID string, userID string) (*model.User, error) {
	return sa.findUserByIDWithContext(context.Background(), orgID, appID, userID)
}

func (sa Adapter) findUserByIDWithContext(context context.Context, orgID string, appID string, userID string) (*model.User, error) {
	filter := bson.D{}
	if len(userID) > 0 {
		filter = bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "user_id", Value: userID},
		}
	}

	var result *model.User
	err := sa.db.users.FindOneWithContext(context, filter, &result, nil)
	if err != nil {
		log.Printf("warning: error while retriving user (%s) - %s", userID, err)
		return nil, err
	}

	return result, err
}

func (sa Adapter) createUserWithContext(context context.Context, orgID string, appID string, userID string, token string, appPlatform *string, appVersion *string, tokenType string) (*model.User, error) {

	now := time.Now().UTC()

	tokenList := []model.DeviceToken{}
	if token != "" {
		tokenList = append(tokenList, model.DeviceToken{
			Token:       token,
			TokenType:   tokenType,
			AppVersion:  appVersion,
			AppPlatform: appPlatform,
			DateCreated: now,
		})
	}
	record := &model.User{
		OrgID:        orgID,
		AppID:        appID,
		ID:           uuid.NewString(),
		UserID:       userID,
		DeviceTokens: tokenList,
		Topics:       []string{},
		DateCreated:  now,
		DateUpdated:  now,
	}

	_, err := sa.db.users.InsertOneWithContext(context, &record)
	if err != nil {
		fmt.Printf("warning: error while inserting token (%s) - %s\n", token, err)
	}

	return record, err
}

func (sa Adapter) addTokenToUserWithContext(ctx context.Context, orgID string, appID string, userID string, token string, appPlatform *string, appVersion *string, tokenType string) error {
	// transaction
	update := bson.D{}

	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "user_id", Value: userID},
	}

	update = bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "date_updated", Value: time.Now().UTC()},
		}},
		primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "firebase_tokens", Value: model.DeviceToken{
			Token:       token,
			TokenType:   tokenType,
			AppVersion:  appVersion,
			AppPlatform: appPlatform,
			DateCreated: time.Now().UTC(),
		}}}},
	}

	_, err := sa.db.users.UpdateOneWithContext(ctx, filter, &update, nil)
	if err != nil {
		fmt.Printf("warning: error while adding token (%s) to user (%s) %s\n", token, userID, err)
		return err
	}

	sa.db.appVersions.InsertOne(map[string]string{
		"org_id": orgID,
		"app_id": appID,
		"name":   *appVersion,
	})

	sa.db.appPlatforms.InsertOne(map[string]string{
		"org_id": orgID,
		"app_id": appID,
		"name":   *appPlatform,
	})
	return nil
}

func (sa Adapter) removeTokenFromUserWithContext(ctx context.Context, orgID string, appID string, token string, userID string, tokenType string) error {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "user_id", Value: userID},
	}

	update := bson.D{}
	update = bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "date_updated", Value: time.Now().UTC()},
		}},
		primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "firebase_tokens", Value: bson.D{primitive.E{Key: "token", Value: token}}}}},
	}

	_, err := sa.db.users.UpdateOneWithContext(ctx, filter, &update, nil)
	if err != nil {
		fmt.Printf("warning: error while removing token (%s) from user (%s) %s\n", token, userID, err)
		return err
	}
	return nil
}

// GetDeviceTokensByRecipients Gets all users mapped to the recipients input list
func (sa Adapter) GetDeviceTokensByRecipients(orgID string, appID string, recipients []model.MessageRecipient, criteriaList []model.RecipientCriteria) ([]string, error) {
	if len(recipients) > 0 {
		innerFilter := []string{}
		for _, recipient := range recipients {
			if recipient.Mute == false {
				innerFilter = append(innerFilter, recipient.UserID)
			}
		}

		filter := bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "user_id", Value: bson.M{"$in": innerFilter}},
		}

		var users []model.User
		err := sa.db.users.Find(filter, &users, nil)
		if err != nil {
			return nil, err
		}

		tokens := []string{}
		for _, tokenMapping := range users {
			if !tokenMapping.NotificationsDisabled {
				for _, token := range tokenMapping.DeviceTokens {
					if len(criteriaList) > 0 {
						include := false
						for _, criteria := range criteriaList {
							if (criteria.AppPlatform == nil || *criteria.AppPlatform == *token.AppPlatform) &&
								(criteria.AppVersion == nil || *criteria.AppVersion == *token.AppVersion) {
								include = true
								break
							}
						}
						if include {
							tokens = append(tokens, token.Token)
						}
					} else {
						tokens = append(tokens, token.Token)
					}
				}
			}
		}

		return tokens, nil
	}
	return nil, fmt.Errorf("empty recient information")
}

// GetUsersByTopicsWithContext Gets all users for topics
func (sa Adapter) GetUsersByTopicsWithContext(ctx context.Context, orgID string, appID string, topics []string) ([]model.User, error) {
	if len(topics) > 0 {
		filter := bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "topics", Value: bson.M{"$in": topics}},
		}

		var tokenMappings []model.User
		err := sa.db.users.FindWithContext(ctx, filter, &tokenMappings, nil)
		if err != nil {
			return nil, err
		}

		// TODO: was this necessary?
		// result := []model.User{}
		// for _, user := range tokenMappings {
		// 	if user.HasTopic(topic) {
		// 		result = append(result, user)
		// 	}
		// }

		return tokenMappings, nil
	}
	return nil, fmt.Errorf("no mapped recipients to %s topics", topics)
}

// GetUsersByRecipientCriteriasWithContext gets users list by list of criteria
func (sa Adapter) GetUsersByRecipientCriteriasWithContext(ctx context.Context, orgID string, appID string, recipientCriterias []model.RecipientCriteria) ([]model.User, error) {
	if len(recipientCriterias) > 0 {
		var users []model.User
		innerFilter := []interface{}{}

		for _, criteria := range recipientCriterias {
			if criteria.AppVersion != nil && len(*criteria.AppVersion) > 0 {
				innerFilter = append(innerFilter, bson.D{primitive.E{Key: "firebase_tokens.app_version", Value: criteria.AppVersion}})
			}
			if criteria.AppPlatform != nil && len(*criteria.AppPlatform) > 0 {
				innerFilter = append(innerFilter, bson.D{primitive.E{Key: "firebase_tokens.app_platform", Value: criteria.AppPlatform}})
			}
		}

		if len(innerFilter) == 0 {
			return nil, fmt.Errorf("no mapped recipients for the input criterias")
		}

		filter := bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "$or", Value: innerFilter},
		}

		err := sa.db.users.FindWithContext(ctx, filter, &users, nil)
		if err != nil {
			return nil, err
		}

		return users, nil
	}
	return nil, fmt.Errorf("no mapped recipients for the input criterias")
}

// UpdateUserByID Updates users notification enabled flag
func (sa Adapter) UpdateUserByID(orgID string, appID string, userID string, notificationsDisabled bool) (*model.User, error) {
	if userID != "" {
		filter := bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "user_id", Value: userID},
		}

		innerUpdate := bson.D{
			primitive.E{Key: "date_updated", Value: time.Now().UTC()},
			primitive.E{Key: "notifications_disabled", Value: notificationsDisabled},
		}

		update := bson.D{
			primitive.E{Key: "$set", Value: innerUpdate},
		}

		_, err := sa.db.users.UpdateOneWithContext(context.Background(), filter, &update, nil)
		if err != nil {
			fmt.Printf("warning: error while updating user record (%s): %s\n", userID, err)
			return nil, err
		}

		return sa.FindUserByID(orgID, appID, userID)
	}
	return nil, nil
}

// DeleteUserWithID Deletes user with ID and all messages
func (sa Adapter) DeleteUserWithID(orgID string, appID string, userID string) error {
	if userID != "" {

		err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
			err := sessionContext.StartTransaction()
			if err != nil {
				log.Printf("error starting a transaction - %s", err)
				abortTransaction(sessionContext)
				return err
			}

			messages, err := sa.FindMessagesRecipientsDeep(orgID, appID, &userID, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			if err != nil {
				fmt.Printf("warning: unable to retrieve messages for user (%s): %s\n", userID, err)
				abortTransaction(sessionContext)
				return err
			}
			if len(messages) > 0 {
				for _, message := range messages {
					err = sa.DeleteUserMessageWithContext(sessionContext, orgID, appID, userID, message.ID)
					if err != nil {
						fmt.Printf("warning: unable to unlink message(%s) for user(%s): %s\n", message.ID, userID, err)
					}

					if *message.Message.CalculatedRecipientsCount == 1 {
						//the message has had only one recipient, so we need to remove the message entity too
						err = sa.DeleteMessagesWithContext(sessionContext, []string{message.ID})
						if err != nil {
							fmt.Printf("warning: unable to delete message(%s): %s\n", message.ID, err)
						}
					}

				}
			}

			filter := bson.D{
				primitive.E{Key: "org_id", Value: orgID},
				primitive.E{Key: "app_id", Value: appID},
				primitive.E{Key: "user_id", Value: userID},
			}
			_, err = sa.db.users.DeleteOneWithContext(sessionContext, filter, nil)
			if err != nil {
				fmt.Printf("warning: error while deleting user record (%s): %s\n", userID, err)
				abortTransaction(sessionContext)
				return err
			}

			if err != nil {
				fmt.Printf("warning: error while delete all messages for user (%s) %s", userID, err)
				abortTransaction(sessionContext)
				return err
			}

			//commit the transaction
			err = sessionContext.CommitTransaction(sessionContext)
			if err != nil {
				fmt.Println(err)
				return err
			}

			return nil
		})
		if err != nil {
			fmt.Printf("warning: error while deleting user record (%s): %s\n", userID, err)
			return err
		}
	}

	return nil
}

// GetMessagesStats counts read/unread and muted/unmuted messages
func (sa *Adapter) GetMessagesStats(userID string) (*model.MessagesStats, error) {
	filter := bson.D{
		primitive.E{Key: "user_id", Value: userID},
	}

	var data []model.MessageRecipient
	err := sa.db.messagesRecipients.Find(filter, &data, nil)
	if err != nil {
		return nil, err
	}
	if data == nil {
		data = make([]model.MessageRecipient, 0)
	}

	totalCount := int64(len(data))
	muted := int64(0)
	unmuted := int64(0)
	read := int64(0)
	unread := int64(0)
	unreadUnmute := int64(0)

	for _, messRec := range data {
		if messRec.Read {
			read++
		} else {
			unread++
		}

		if messRec.Mute {
			muted++
		} else {
			unmuted++
		}
		if messRec.Read == false && messRec.Mute == false {
			unreadUnmute++
		}
	}

	stats := model.MessagesStats{TotalCount: &totalCount, Muted: &muted,
		Unmuted: &unmuted, Read: &read, Unread: &unread, UnreadUnmute: &unreadUnmute}
	return &stats, nil
}

// SubscribeToTopic subscribes the token to a topic
func (sa Adapter) SubscribeToTopic(orgID string, appID string, token string, userID string, topic string) error {
	record, err := sa.FindUserByID(orgID, appID, userID)
	if err == nil && record != nil {
		if err == nil && record != nil && !record.HasTopic(topic) {
			filter := bson.D{
				primitive.E{Key: "org_id", Value: orgID},
				primitive.E{Key: "app_id", Value: appID},
				primitive.E{Key: "user_id", Value: record.UserID},
			}
			update := bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "date_updated", Value: time.Now().UTC()},
				}},
				primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "topics", Value: topic}}},
			}
			_, err = sa.db.users.UpdateOne(filter, update, nil)
			if err == nil {
				var topicRecord *model.Topic
				topicRecord, _ = sa.GetTopicByName(orgID, appID, topic)
				if topicRecord == nil {
					sa.InsertTopic(&model.Topic{OrgID: orgID, AppID: appID, Name: topic}) // just try to append within the topics collection
				}
			}
		}
	}
	return err
}

// UnsubscribeToTopic unsubscribes the token from a topic
func (sa Adapter) UnsubscribeToTopic(orgID string, appID string, token string, userID string, topic string) error {
	record, err := sa.FindUserByID(orgID, appID, userID)
	if err == nil && record != nil {
		if err == nil && record != nil && record.HasTopic(topic) {
			filter := bson.D{
				primitive.E{Key: "org_id", Value: orgID},
				primitive.E{Key: "app_id", Value: appID},
				primitive.E{Key: "user_id", Value: record.UserID},
			}
			update := bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "date_updated", Value: time.Now().UTC()},
				}},
				primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "topics", Value: topic}}},
			}
			_, err = sa.db.users.UpdateOne(filter, update, nil)
			if err == nil {
				var topicRecord *model.Topic
				topicRecord, _ = sa.GetTopicByName(orgID, appID, topic)
				if topicRecord == nil {
					sa.InsertTopic(&model.Topic{OrgID: orgID, AppID: appID, Name: topic}) // just try to append within the topics collection in case it's missing
				}
			}
		}
	}
	return err
}

// GetTopics gets all topics
func (sa Adapter) GetTopics(orgID string, appID string) ([]model.Topic, error) {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
	}
	var result []model.Topic

	err := sa.db.topics.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetTopicByName appends a new topic within the topics collection
func (sa Adapter) GetTopicByName(orgID string, appID string, name string) (*model.Topic, error) {
	if name != "" {
		filter := bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "_id", Value: name},
		}
		var topic model.Topic
		err := sa.db.topics.FindOne(filter, &topic, nil)
		if err == nil {
			return &topic, nil
		}
		fmt.Printf("warning: error while retriving topic (%s) - %s\n", name, err)
		return nil, err
	}
	return nil, nil
}

// InsertTopic appends a new topic within the topics collection
func (sa Adapter) InsertTopic(topic *model.Topic) (*model.Topic, error) {
	if topic.Name != "" {
		now := time.Now().UTC()
		topic.DateUpdated = now
		topic.DateCreated = now

		_, err := sa.db.topics.InsertOne(&topic)
		if err != nil {
			fmt.Printf("warning: error while store topic (%s) - %s\n", topic.Name, err)
			return nil, err
		}
	}

	return topic, nil
}

// UpdateTopic updates a topic (for now only description is updatable)
func (sa Adapter) UpdateTopic(topic *model.Topic) (*model.Topic, error) {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: topic.OrgID},
		primitive.E{Key: "app_id", Value: topic.AppID},
		primitive.E{Key: "_id", Value: topic.Name},
	}

	now := time.Now().UTC()
	topic.DateUpdated = now

	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "description", Value: topic.Description},
			primitive.E{Key: "date_updated", Value: topic.DateUpdated},
		}},
	}

	_, err := sa.db.topics.UpdateOne(filter, &update, nil)
	if err != nil {
		fmt.Printf("warning: error while update topic (%s) - %s\n", topic.Name, err)
		return nil, err
	}

	return topic, err
}

// FindMessagesRecipients finds messages recipients
func (sa Adapter) FindMessagesRecipients(orgID string, appID string, messageID string, userID string) ([]model.MessageRecipient, error) {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "message_id", Value: messageID},
		primitive.E{Key: "user_id", Value: userID},
	}

	var data []model.MessageRecipient
	err := sa.db.messagesRecipients.Find(filter, &data, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FindMessagesRecipientsByMessageAndUsers finds messages recipients by message and users
func (sa Adapter) FindMessagesRecipientsByMessageAndUsers(messageID string, usersIDs []string) ([]model.MessageRecipient, error) {
	filter := bson.D{
		primitive.E{Key: "message_id", Value: messageID},
		primitive.E{Key: "user_id", Value: bson.M{"$in": usersIDs}},
	}

	var data []model.MessageRecipient
	err := sa.db.messagesRecipients.Find(filter, &data, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FindMessagesRecipientsByMessages finds messages recipients by messages
func (sa Adapter) FindMessagesRecipientsByMessages(messagesIDs []string) ([]model.MessageRecipient, error) {
	filter := bson.D{
		primitive.E{Key: "message_id", Value: bson.M{"$in": messagesIDs}},
	}

	var data []model.MessageRecipient
	err := sa.db.messagesRecipients.Find(filter, &data, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FindMessagesRecipientsDeep finds messages recipients join with messages
func (sa Adapter) FindMessagesRecipientsDeep(orgID string, appID string, userID *string, read *bool, mute *bool,
	messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string,
	offset *int64, limit *int64, order *string) ([]model.MessageRecipient, error) {

	type recipientJoinMessage struct {
		//message
		Priority                  int                       `bson:"priority"`
		Subject                   string                    `bson:"subject"`
		Sender                    model.Sender              `bson:"sender"`
		Body                      string                    `bson:"body"`
		Data                      map[string]string         `bson:"data"`
		Recipients                []model.MessageRecipient  `bson:"recipients"`
		RecipientsCriteriaList    []model.RecipientCriteria `bson:"recipients_criteria_list"`
		RecipientAccountCriteria  map[string]interface{}    `bson:"recipient_account_criteria"`
		Topic                     *string                   `bson:"topic"`
		Topics                    []string                  `bson:"topics"`
		CalculatedRecipientsCount *int                      `bson:"calculated_recipients_count"`
		DateCreated               *time.Time                `bson:"date_created"`
		DateUpdated               *time.Time                `bson:"date_updated"`
		Time                      time.Time                 `bson:"time"`

		//recipient
		OrgID     string `bson:"org_id"`
		AppID     string `bson:"app_id"`
		ID        string `bson:"_id"`
		UserID    string `bson:"user_id"`
		MessageID string `bson:"message_id"`
		Mute      bool   `bson:"mute"`
		Read      bool   `bson:"read"`
	}

	pipeline := []bson.M{
		{"$lookup": bson.M{
			"from":         "messages",
			"localField":   "message_id",
			"foreignField": "_id",
			"as":           "message",
		}},
		{"$unwind": "$message"},
		{"$project": bson.M{"org_id": 1, "app_id": 1, "_id": 1,
			"user_id": 1, "message_id": 1, "mute": 1, "read": 1, "time": "$message.time",
			"priority": "$message.priority", "subject": "$message.subject", "sender": "$message.sender",
			"body": "$message.body", "data": "$message.data", "recipients": "$message.recipients",
			"recipients_criteria_list": "$message.recipients_criteria_list", "recipient_account_criteria": "$message.recipient_account_criteria",
			"topic": "$message.topic", "topics": "$message.topics", "calculated_recipients_count": "$message.calculated_recipients_count",
			"date_created": "$message.date_created", "date_updated": "$message.date_updated"}},
		{"$match": bson.M{"org_id": orgID}},
		{"$match": bson.M{"app_id": appID}},
	}

	if userID != nil && len(*userID) > 0 {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"user_id": *userID}})
	}

	if read != nil {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"read": *read}})
	}

	if mute != nil {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"mute": *mute}})
	}

	if len(messageIDs) > 0 {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"message_id": bson.M{"$in": messageIDs}}})
	}

	if filterTopic != nil {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"topic": *filterTopic}})
	}

	pipeline = append(pipeline, bson.M{"$match": bson.M{"time": bson.M{"$lte": time.Now()}}})

	if startDateEpoch != nil {
		seconds := *startDateEpoch / 1000
		timeValue := time.Unix(seconds, 0)
		pipeline = append(pipeline, bson.M{"$match": bson.M{"time": bson.D{primitive.E{Key: "$gte", Value: &timeValue}}}})
	}
	if endDateEpoch != nil {
		seconds := *endDateEpoch / 1000
		timeValue := time.Unix(seconds, 0)
		pipeline = append(pipeline, bson.M{"$match": bson.M{"time": bson.D{primitive.E{Key: "$lte", Value: &timeValue}}}})
	}

	if order != nil && *order == "asc" {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"time": 1}})
	} else {
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"time": -1}})
	}

	if limit != nil {
		//calculate real limit
		offsetValue := utils.GetInt64Value(offset)
		calculatedLimit := offsetValue + *limit
		pipeline = append(pipeline, bson.M{"$limit": calculatedLimit})
	}
	if offset != nil {
		pipeline = append(pipeline, bson.M{"$skip": *offset})
	}

	var items []recipientJoinMessage
	err := sa.db.messagesRecipients.Aggregate(pipeline, &items, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "message", nil, err)
	}

	result := make([]model.MessageRecipient, len(items))
	for i, item := range items {

		message := model.Message{OrgID: item.OrgID, AppID: item.AppID, ID: item.MessageID,
			Priority: item.Priority, Subject: item.Subject,
			Sender: item.Sender, Body: item.Body, Data: item.Data, Recipients: item.Recipients,
			RecipientsCriteriaList: item.RecipientsCriteriaList, RecipientAccountCriteria: item.RecipientAccountCriteria,
			Topic: item.Topic, Topics: item.Topics, CalculatedRecipientsCount: item.CalculatedRecipientsCount, DateCreated: item.DateCreated,
			DateUpdated: item.DateUpdated, Time: item.Time}

		recipient := model.MessageRecipient{OrgID: item.OrgID, AppID: item.AppID,
			ID: item.ID, UserID: item.UserID, MessageID: item.MessageID, Mute: item.Mute,
			Read: item.Read, Message: message}
		result[i] = recipient
	}

	return result, nil
}

// InsertMessagesRecipientsWithContext inserts messages recipients
func (sa Adapter) InsertMessagesRecipientsWithContext(ctx context.Context, items []model.MessageRecipient) error {
	if len(items) == 0 {
		return nil
	}

	data := make([]interface{}, len(items))
	for i, p := range items {
		data[i] = p
	}

	res, err := sa.db.messagesRecipients.InsertManyWithContext(ctx, data, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "messages recipients", nil, err)
	}

	if len(res.InsertedIDs) != len(items) {
		return errors.ErrorAction(logutils.ActionInsert, "messages recipients", &logutils.FieldArgs{"inserted": len(res.InsertedIDs), "expected": len(items)})
	}

	return nil
}

// DeleteMessagesRecipientsForIDsWithContext deletes messages recipients for ids
func (sa Adapter) DeleteMessagesRecipientsForIDsWithContext(ctx context.Context, ids []string) error {
	filter := bson.D{primitive.E{Key: "_id", Value: bson.M{"$in": ids}}}

	_, err := sa.db.messagesRecipients.DeleteManyWithContext(ctx, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "message recipient", nil, err)
	}
	return nil
}

// DeleteMessagesRecipientsForMessagesWithContext deletes messages recipients for messages
func (sa Adapter) DeleteMessagesRecipientsForMessagesWithContext(ctx context.Context, messagesIDs []string) error {
	filter := bson.D{primitive.E{Key: "message_id", Value: bson.M{"$in": messagesIDs}}}

	_, err := sa.db.messagesRecipients.DeleteManyWithContext(ctx, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "message recipient", nil, err)
	}
	return nil
}

// FindMessagesWithContext finds messages by ids using context
func (sa Adapter) FindMessagesWithContext(ctx context.Context, ids []string) ([]model.Message, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: bson.M{"$in": ids}}}

	var messageArr []model.Message
	err := sa.db.messages.FindWithContext(ctx, filter, &messageArr, nil)
	if err != nil {
		return nil, err
	}

	return messageArr, nil
}

// FindMessagesByParams finds messages by params
func (sa Adapter) FindMessagesByParams(orgID string, appID string, senderType string, senderAccountID *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "sender.type", Value: senderType},
	}
	//sender account id
	if senderAccountID != nil {
		filter = append(filter, primitive.E{Key: "sender.user.user_id", Value: *senderAccountID})
	}

	findOptions := options.Find()
	//limit
	limitValue := int64(50) //by default - 50
	if limit != nil {
		limitValue = int64(*limit)
	}
	findOptions.SetLimit(limitValue)

	//offset
	if offset != nil {
		findOptions.SetSkip(int64(*offset))
	}
	//sort
	sortValue := -1 //by default -  "asc"
	if order != nil && *order == "desc" {
		sortValue = 1
	}
	findOptions.SetSort(bson.D{primitive.E{Key: "date_created", Value: sortValue}})

	var messages []model.Message
	err := sa.db.messages.Find(filter, &messages, findOptions)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// GetMessage gets a message by id
func (sa Adapter) GetMessage(orgID string, appID string, ID string) (*model.Message, error) {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "_id", Value: ID},
	}

	var message *model.Message
	err := sa.db.messages.FindOne(filter, &message, nil)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// CreateMessageWithContext creates a new message.
func (sa Adapter) CreateMessageWithContext(ctx context.Context, message model.Message) (*model.Message, error) {
	if len(message.ID) == 0 {
		id := uuid.New().String()
		message.ID = id
	}
	now := time.Now().UTC()
	message.DateUpdated = &now
	message.DateCreated = &now

	_, err := sa.db.messages.InsertOneWithContext(ctx, &message)
	if err != nil {
		fmt.Printf("warning: error while store message (%s) - %s", message.ID, err)
		return nil, err
	}

	return &message, nil
}

// InsertMessagesWithContext inserts messages.
func (sa Adapter) InsertMessagesWithContext(ctx context.Context, messages []model.Message) error {
	data := make([]interface{}, len(messages))
	for i, p := range messages {
		data[i] = p
	}

	res, err := sa.db.messages.InsertManyWithContext(ctx, data, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "messagess", nil, err)
	}

	if len(res.InsertedIDs) != len(messages) {
		return errors.ErrorAction(logutils.ActionInsert, "messages", &logutils.FieldArgs{"inserted": len(res.InsertedIDs), "expected": len(messages)})
	}

	return nil
}

// UpdateMessage updates a message
func (sa Adapter) UpdateMessage(message *model.Message) (*model.Message, error) {
	if message != nil {
		persistedMessage, err := sa.GetMessage(message.OrgID, message.AppID, message.ID)
		if err != nil || persistedMessage == nil {
			return nil, fmt.Errorf("Message with id (%s) not found: %w", message.ID, err)
		}

		filter := bson.D{
			primitive.E{Key: "org_id", Value: message.OrgID},
			primitive.E{Key: "app_id", Value: message.AppID},
			primitive.E{Key: "_id", Value: message.ID},
		}

		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "priority", Value: message.Priority},
				primitive.E{Key: "topic", Value: message.Topic},
				primitive.E{Key: "subject", Value: message.Subject},
				primitive.E{Key: "body", Value: message.Body},
				primitive.E{Key: "date_updated", Value: time.Now().UTC()},
				primitive.E{Key: "topics", Value: message.Topics},
			}},
		}

		_, err = sa.db.messages.UpdateOne(filter, update, nil)
		if err != nil {
			fmt.Printf("warning: error while update message (%s) - %s", message.ID, err)
			return nil, err
		}
	}

	return message, nil
}

// DeleteUserMessageWithContext removes the desired user from the recipients list
func (sa Adapter) DeleteUserMessageWithContext(ctx context.Context, orgID string, appID string, userID string, messageID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	persistedMessage, err := sa.GetMessage(orgID, appID, messageID)
	if err != nil || persistedMessage == nil {
		return fmt.Errorf("message with id (%s) not found: %s", messageID, err)
	}

	//remove the messages recipients records
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "message_id", Value: messageID},
		primitive.E{Key: "user_id", Value: userID}}

	_, err = sa.db.messagesRecipients.DeleteManyWithContext(ctx, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "message recipient",
			&logutils.FieldArgs{"user_id": userID, "message_id": messageID}, err)
	}
	return nil
}

// DeleteMessagesWithContext deletes messages by ids
func (sa Adapter) DeleteMessagesWithContext(ctx context.Context, ids []string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	filter := bson.D{primitive.E{Key: "_id", Value: bson.M{"$in": ids}}}
	_, err := sa.db.messages.DeleteManyWithContext(ctx, filter, nil)
	if err != nil {
		fmt.Printf("warning: error while delete messages - %s", err)
		return err
	}

	return nil
}

// UpdateUnreadMessage updates a unread message in the recipients to read
func (sa Adapter) UpdateUnreadMessage(ctx context.Context, orgID string, appID string, ID string, userID string) (*model.Message, error) {
	read := true
	filter := bson.D{primitive.E{Key: "message_id", Value: ID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "read", Value: read},
		}},
	}
	_, err := sa.db.messagesRecipients.UpdateOneWithContext(ctx, filter, update, nil)
	if err != nil {
		fmt.Println("warning: error while updating massage", ID, userID, err)
		return nil, err
	}
	return nil, nil
}

// UpdateAllUserMessagesRead Update all user messages as read or as unread
func (sa Adapter) UpdateAllUserMessagesRead(ctx context.Context, orgID string, appID string, userID string, read bool) error {
	filter := bson.D{
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "read", Value: read},
		}},
	}
	_, err := sa.db.messagesRecipients.UpdateManyWithContext(ctx, filter, update, nil)
	if err != nil {
		fmt.Println("warning: error while read/unread all user messages", userID, err)
		return err
	}
	return nil
}

// GetAllAppVersions gets all registered versions
func (sa Adapter) GetAllAppVersions(orgID string, appID string) ([]model.AppVersion, error) {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
	}

	var versions []model.AppVersion
	err := sa.db.appVersions.Find(filter, &versions, nil)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

// GetAllAppPlatforms gets all registered platforms
func (sa Adapter) GetAllAppPlatforms(orgID string, appID string) ([]model.AppPlatform, error) {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
	}

	var platforms []model.AppPlatform
	err := sa.db.appPlatforms.Find(filter, &platforms, nil)
	if err != nil {
		return nil, err
	}

	return platforms, nil
}

// InsertQueueDataItemsWithContext inserts queue data items
func (sa Adapter) InsertQueueDataItemsWithContext(ctx context.Context, items []model.QueueItem) error {
	if len(items) == 0 {
		return nil
	}

	data := make([]interface{}, len(items))
	for i, p := range items {
		data[i] = p
	}

	res, err := sa.db.queueData.InsertManyWithContext(ctx, data, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "queue data items", nil, err)
	}

	if len(res.InsertedIDs) != len(items) {
		return errors.ErrorAction(logutils.ActionInsert, "queue data items", &logutils.FieldArgs{"inserted": len(res.InsertedIDs), "expected": len(items)})
	}

	return nil
}

// LoadQueueWithContext loads the queue object
func (sa Adapter) LoadQueueWithContext(ctx context.Context) (*model.Queue, error) {
	filter := bson.D{}

	var queue []model.Queue
	err := sa.db.queue.FindWithContext(ctx, filter, &queue, nil)
	if err != nil {
		return nil, err
	}

	if len(queue) == 0 {
		return nil, nil
	}

	res := queue[0] //we support only one record
	return &res, nil
}

// SaveQueueWithContext saves queue with context
func (sa *Adapter) SaveQueueWithContext(ctx context.Context, queue model.Queue) error {
	filter := bson.D{primitive.E{Key: "_id", Value: queue.ID}}
	err := sa.db.queue.ReplaceOneWithContext(ctx, filter, queue, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "queue", &logutils.FieldArgs{"_id": queue.ID}, err)
	}
	return nil
}

// SaveQueue saves queue
func (sa *Adapter) SaveQueue(queue model.Queue) error {
	filter := bson.D{primitive.E{Key: "_id", Value: queue.ID}}
	err := sa.db.queue.ReplaceOne(filter, queue, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "queue", &logutils.FieldArgs{"_id": queue.ID}, err)
	}
	return nil
}

// FindQueueData finds queue data
func (sa *Adapter) FindQueueData(time *time.Time, limit int) ([]model.QueueItem, error) {
	filter := bson.D{}

	//time
	if time != nil {
		filter = append(filter, primitive.E{Key: "time", Value: bson.M{"$lte": time}})
	}

	//set limit
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))

	//set sort
	findOptions.SetSort(bson.D{primitive.E{Key: "time", Value: 1}, primitive.E{Key: "priority", Value: 1}})

	var result []model.QueueItem
	err := sa.db.queueData.Find(filter, &result, findOptions)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "queue data", nil, err)
	}
	return result, nil
}

// DeleteQueueData removes queue data
func (sa *Adapter) DeleteQueueData(ids []string) error {
	filter := bson.D{primitive.E{Key: "_id", Value: bson.M{"$in": ids}}}
	_, err := sa.db.queueData.DeleteMany(filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "queue data", &logutils.FieldArgs{"ids": ids}, err)
	}
	return nil
}

// DeleteQueueDataForMessagesWithContext removes queue data items for messages
func (sa *Adapter) DeleteQueueDataForMessagesWithContext(ctx context.Context, messagesIDs []string) error {
	filter := bson.D{primitive.E{Key: "message_id", Value: bson.M{"$in": messagesIDs}}}

	_, err := sa.db.queueData.DeleteManyWithContext(ctx, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "queue data", nil, err)
	}
	return nil
}

// DeleteQueueDataForRecipientsWithContext removes queue data items for recepients
func (sa *Adapter) DeleteQueueDataForRecipientsWithContext(ctx context.Context, recipientsIDs []string) error {
	filter := bson.D{primitive.E{Key: "message_recipient_id", Value: bson.M{"$in": recipientsIDs}}}

	_, err := sa.db.queueData.DeleteManyWithContext(ctx, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "queue data", nil, err)
	}
	return nil
}

// StoreDeviceToken stores device token
func (sa Adapter) StoreDeviceToken(orgID string, appID string, tokenInfo *model.TokenInfo, userID string) error {
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		userRecord, _ := sa.findUserByTokenWithContext(sessionContext, orgID, appID, tokenInfo.Token)
		if userRecord == nil {
			existingUser, _ := sa.findUserByIDWithContext(sessionContext, orgID, appID, userID)
			if existingUser != nil {
				err = sa.addTokenToUserWithContext(sessionContext, orgID, appID, userID, tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion, tokenInfo.TokenType)
			} else {
				_, err = sa.createUserWithContext(sessionContext, orgID, appID, userID, tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion, tokenInfo.TokenType)
			}
		} else if userRecord.UserID != userID {
			err = sa.removeTokenFromUserWithContext(sessionContext, orgID, appID, tokenInfo.Token, userRecord.UserID, tokenInfo.TokenType)
			if err != nil {
				fmt.Printf("error while unlinking token (%s) from user (%s)- %s\n", tokenInfo.Token, userRecord.UserID, err)
				return err
			}

			existingUser, _ := sa.findUserByIDWithContext(sessionContext, orgID, appID, userID)
			if existingUser != nil {
				err = sa.addTokenToUserWithContext(sessionContext, orgID, appID, userID, tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion, tokenInfo.TokenType)
			} else {
				_, err = sa.createUserWithContext(sessionContext, orgID, appID, userID, tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion, tokenInfo.TokenType)
			}
			if err != nil {
				fmt.Printf("error while linking token (%s) from user (%s)- %s\n", tokenInfo.Token, userID, err)
				return err
			}
		}

		if err != nil {
			fmt.Printf("error while storing token (%s) to user (%s) %s\n", tokenInfo.Token, userID, err)
			abortTransaction(sessionContext)
			return err
		}

		//commit the transaction
		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			abortTransaction(sessionContext)
			fmt.Println(err)
			return err
		}
		return nil
	})

	return err
}

// FindConfig finds the config for the specified type, appID, and orgID
func (sa *Adapter) FindConfig(configType string, appID string, orgID string) (*model.Configs, error) {
	return sa.getCachedConfig("", configType, appID, orgID)
}

// FindConfigByID finds the config for the specified ID
func (sa *Adapter) FindConfigByID(id string) (*model.Configs, error) {
	return sa.getCachedConfig(id, "", "", "")
}

// FindConfigs finds all configs for the specified type
func (sa *Adapter) FindConfigs(configType *string) ([]model.Configs, error) {
	return sa.getCachedConfigs(configType)
}

func (sa *Adapter) setCachedConfigs(configs []model.Configs) {
	sa.configsLock.Lock()
	defer sa.configsLock.Unlock()

	sa.cachedConfigs = &syncmap.Map{}

	for _, config := range configs {
		var err error
		switch config.Type {
		case model.ConfigTypeEnv:
			err = parseConfigsData[model.EnvConfigData](&config)
		default:
			err = parseConfigsData[map[string]interface{}](&config)
		}
		if err != nil {
			sa.db.logger.Warn(err.Error())
		}
		sa.cachedConfigs.Store(config.ID, config)
		sa.cachedConfigs.Store(fmt.Sprintf("%s_%s_%s", config.Type, config.AppID, config.OrgID), config)
	}
}

func parseConfigsData[T model.ConfigData](config *model.Configs) error {
	bsonBytes, err := bson.Marshal(config.Data)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUnmarshal, model.TypeConfig, nil, err)
	}

	var data T
	err = bson.Unmarshal(bsonBytes, &data)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUnmarshal, model.TypeConfigData, &logutils.FieldArgs{"type": config.Type}, err)
	}

	config.Data = data
	return nil
}

func (sa *Adapter) getCachedConfig(id string, configType string, appID string, orgID string) (*model.Configs, error) {
	sa.configsLock.RLock()
	defer sa.configsLock.RUnlock()

	var item any
	var errArgs logutils.FieldArgs
	if id != "" {
		errArgs = logutils.FieldArgs{"id": id}
		item, _ = sa.cachedConfigs.Load(id)
	} else {
		errArgs = logutils.FieldArgs{"type": configType, "app_id": appID, "org_id": orgID}
		item, _ = sa.cachedConfigs.Load(fmt.Sprintf("%s_%s_%s", configType, appID, orgID))
	}

	if item != nil {
		config, ok := item.(model.Configs)
		if !ok {
			return nil, errors.ErrorAction(logutils.ActionCast, model.TypeConfig, &errArgs)
		}
		return &config, nil
	}
	return nil, nil
}

func (sa *Adapter) getCachedConfigs(configType *string) ([]model.Configs, error) {
	sa.configsLock.RLock()
	defer sa.configsLock.RUnlock()

	var err error
	configList := make([]model.Configs, 0)
	sa.cachedConfigs.Range(func(key, item interface{}) bool {
		keyStr, ok := key.(string)
		if !ok || item == nil {
			return false
		}
		if !strings.Contains(keyStr, "_") {
			return true
		}

		config, ok := item.(model.Configs)
		if !ok {
			err = errors.ErrorAction(logutils.ActionCast, model.TypeConfig, &logutils.FieldArgs{"key": key})
			return false
		}

		if configType == nil || strings.HasPrefix(keyStr, fmt.Sprintf("%s_", *configType)) {
			configList = append(configList, config)
		}

		return true
	})

	return configList, err
}

// loadConfigs loads configs
func (sa *Adapter) loadConfigs() ([]model.Configs, error) {
	filter := bson.M{}

	var configs []model.Configs
	err := sa.db.configs.Find(filter, &configs, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}

	return configs, nil
}

// InsertConfig inserts a new config
func (sa *Adapter) InsertConfig(config model.Configs) error {
	_, err := sa.db.configs.InsertOne(config)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeConfig, nil, err)
	}

	return nil
}

// UpdateConfig updates an existing config
func (sa *Adapter) UpdateConfig(config model.Configs) error {
	filter := bson.M{"_id": config.ID}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "type", Value: config.Type},
			primitive.E{Key: "app_id", Value: config.AppID},
			primitive.E{Key: "org_id", Value: config.OrgID},
			primitive.E{Key: "system", Value: config.System},
			primitive.E{Key: "data", Value: config.Data},
			primitive.E{Key: "date_updated", Value: config.DateUpdated},
		}},
	}
	_, err := sa.db.configs.UpdateOne(filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeConfig, &logutils.FieldArgs{"id": config.ID}, err)
	}

	return nil
}

// DeleteConfig deletes a configuration from storage
func (sa *Adapter) DeleteConfig(id string) error {
	delFilter := bson.M{"_id": id}
	_, err := sa.db.configs.DeleteMany(delFilter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeConfig, &logutils.FieldArgs{"id": id}, err)
	}

	return nil
}

// cacheConfigs caches the configs from the DB
func (sa *Adapter) cacheConfigs() error {
	sa.db.logger.Info("cacheConfigs...")

	configs, err := sa.loadConfigs()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionLoad, model.TypeConfig, nil, err)
	}

	sa.setCachedConfigs(configs)

	return nil
}

func abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error on aborting a transaction - %s", err)
	}
}

func (sa *Adapter) abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error aborting a transaction - %s", err)
	}
}

// Listener represents storage listener
type Listener interface {
	OnFirebaseConfigurationsUpdated()
}

// TransactionContext represents storage transaction interface
type TransactionContext interface {
	mongo.SessionContext
}
