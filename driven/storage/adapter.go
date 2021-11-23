/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, AppVersion 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package storage

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"notifications/core/model"
	"strconv"
	"time"
)

// Adapter implements the Storage interface
type Adapter struct {
	db *database
}

// Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
	return err
}

// NewStorageAdapter creates a new storage adapter instance
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		log.Println("Set default timeout - 2000")
		timeout = 2000
	}
	timeoutMS := time.Millisecond * time.Duration(timeout)

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS}
	return &Adapter{db: db}
}

// FindUserByToken finds firebase token
func (sa Adapter) FindUserByToken(token string) (*model.User, error) {
	return sa.findUserByTokenWithContext(context.Background(), token)
}

func (sa Adapter) findUserByTokenWithContext(context context.Context, token string) (*model.User, error) {
	filter := bson.D{}
	if len(token) > 0 {
		filter = bson.D{
			primitive.E{Key: "firebase_tokens", Value: bson.D{primitive.E{Key: "$elemMatch", Value: bson.D{primitive.E{Key: "token", Value: token}}}}},
		}
	}

	var result *model.User
	err := sa.db.users.FindOneWithContext(context, filter, &result, nil)
	if err != nil {
		log.Printf("warning: error while retriving token (%s) - %s", token, err)
	}

	return result, err
}

// FindUserByID finds user by id
func (sa Adapter) FindUserByID(userID string) (*model.User, error) {
	return sa.findUserByIDWithContext(context.Background(), userID)
}

func (sa Adapter) findUserByIDWithContext(context context.Context, userID string) (*model.User, error) {
	filter := bson.D{}
	if len(userID) > 0 {
		filter = bson.D{
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

// StoreFirebaseToken stores firebase token
func (sa Adapter) StoreFirebaseToken(tokenInfo *model.TokenInfo, user *model.CoreToken) error {

	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		// Remove previous token no matter on with user is linked
		if tokenInfo.PreviousToken != nil {
			existingUser, _ := sa.findUserByTokenWithContext(sessionContext, *tokenInfo.PreviousToken)
			if existingUser != nil {
				err = sa.removeTokenFromUserWithContext(sessionContext, *tokenInfo.PreviousToken, existingUser.UserID)
				if err != nil {
					fmt.Printf("error while removing the previous token (%s) from user (%s)- %s\n", *tokenInfo.PreviousToken, *user.UserID, err)
					return err
				}
			}
		}

		userRecord, _ := sa.findUserByTokenWithContext(sessionContext, *tokenInfo.Token)
		if userRecord == nil {
			if user.UserID != nil {
				existingUser, _ := sa.findUserByIDWithContext(sessionContext, *user.UserID)
				if existingUser != nil {
					err = sa.addTokenToUserWithContext(sessionContext, user.UserID, *tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion)
				} else {
					_, err = sa.createUserWithContext(sessionContext, user.UserID, *tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion)
				}
			}
		} else if userRecord.UserID != nil && userRecord.UserID != user.UserID {
			err = sa.removeTokenFromUserWithContext(sessionContext, *tokenInfo.Token, userRecord.UserID)
			if err != nil {
				fmt.Printf("error while unlinking token (%s) from user (%s)- %s\n", *tokenInfo.Token, *userRecord.UserID, err)
				return err
			}

			existingUser, _ := sa.findUserByIDWithContext(sessionContext, *user.UserID)
			if existingUser != nil {
				err = sa.addTokenToUserWithContext(sessionContext, user.UserID, *tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion)
			} else {
				_, err = sa.createUserWithContext(sessionContext, user.UserID, *tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion)
			}
			if err != nil {
				fmt.Printf("error while linking token (%s) from user (%s)- %s\n", *tokenInfo.Token, *user.UserID, err)
				return err
			}
		}

		if err != nil {
			fmt.Printf("error while storing token (%s) to user (%s) %s\n", *tokenInfo.Token, *user.UserID, err)
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

func (sa Adapter) createUserWithContext(context context.Context, userID *string, token string, appPlatform *string, appVersion *string) (*model.User, error) {

	now := time.Now().UTC()

	tokenList := []model.FirebaseToken{}
	if token != "" {
		tokenList = append(tokenList, model.FirebaseToken{
			Token:       token,
			AppVersion:  appVersion,
			AppPlatform: appPlatform,
			DateCreated: now,
		})
	}
	record := &model.User{
		ID:             uuid.NewString(),
		UserID:         userID,
		FirebaseTokens: tokenList,
		Topics:         []string{},
		DateCreated:    now,
		DateUpdated:    now,
	}

	_, err := sa.db.users.InsertOneWithContext(context, &record)
	if err != nil {
		fmt.Printf("warning: error while inserting token (%s) - %s\n", token, err)
	}

	return record, err
}

func (sa Adapter) addTokenToUserWithContext(ctx context.Context, userID *string, token string, appPlatform *string, appVersion *string) error {
	if userID != nil {
		// transaction
		filter := bson.D{primitive.E{Key: "user_id", Value: userID}}

		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now().UTC()},
			}},
			primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "firebase_tokens", Value: model.FirebaseToken{
				Token:       token,
				AppVersion:  appVersion,
				AppPlatform: appPlatform,
				DateCreated: time.Now().UTC(),
			}}}},
		}

		_, err := sa.db.users.UpdateOneWithContext(ctx, filter, &update, nil)
		if err != nil {
			fmt.Printf("warning: error while adding token (%s) to user (%s) %s\n", token, *userID, err)
			return err
		}

		sa.db.appVersions.InsertOne(map[string]string{
			"name": *appVersion,
		})

		sa.db.appPlatforms.InsertOne(map[string]string{
			"name": *appPlatform,
		})
	}
	return nil
}

func (sa Adapter) removeTokenFromUserWithContext(ctx context.Context, token string, userID *string) error {
	if userID != nil {
		filter := bson.D{primitive.E{Key: "user_id", Value: userID}}

		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "date_updated", Value: time.Now().UTC()},
			}},
			primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "firebase_tokens", Value: bson.D{primitive.E{Key: "token", Value: token}}}}},
		}

		_, err := sa.db.users.UpdateOneWithContext(ctx, filter, &update, nil)
		if err != nil {
			fmt.Printf("warning: error while removing token (%s) from user (%s) %s\n", token, *userID, err)
			return err
		}
	}
	return nil
}

// GetFirebaseTokensByRecipients Gets all users mapped to the recipients input list
func (sa Adapter) GetFirebaseTokensByRecipients(recipients []model.Recipient, topic *string) ([]string, error) {
	if len(recipients) > 0 {
		innerFilter := []string{}
		for _, recipient := range recipients {
			if recipient.UserID != nil {
				innerFilter = append(innerFilter, *recipient.UserID)
			}
		}

		filter := bson.D{
			primitive.E{Key: "user_id", Value: bson.M{"$in": innerFilter}},
		}

		var users []model.User
		err := sa.db.users.Find(filter, &users, nil)
		if err != nil {
			return nil, err
		}

		tokens := []string{}
		for _, tokenMapping := range users {
			if !tokenMapping.NotificationsDisabled && (topic == nil || tokenMapping.HasTopic(*topic)) {
				for _, token := range tokenMapping.FirebaseTokens {
					tokens = append(tokens, token.Token)
				}
			}
		}

		return tokens, nil
	}
	return nil, fmt.Errorf("empty recient information")
}

// GetRecipientsByTopic Gets all users recipients by topic
func (sa Adapter) GetRecipientsByTopic(topic string) ([]model.Recipient, error) {
	if len(topic) > 0 {
		filter := bson.D{primitive.E{Key: "topics", Value: topic}}

		var tokenMappings []model.User
		err := sa.db.users.Find(filter, &tokenMappings, nil)
		if err != nil {
			return nil, err
		}

		recipients := []model.Recipient{}
		for _, user := range tokenMappings {
			if user.HasTopic(topic) {
				recipients = append(recipients, model.Recipient{
					UserID:               user.UserID,
					NotificationDisabled: user.NotificationsDisabled,
				})
			}
		}

		return recipients, nil
	}
	return nil, fmt.Errorf("no mapped recipients to %s topic", topic)
}

// GetRecipientsByRecipientCriterias gets recipients list by list of criteria
func (sa Adapter) GetRecipientsByRecipientCriterias(recipientCriterias []model.RecipientCriteria) ([]model.Recipient, error) {
	if len(recipientCriterias) > 0 {
		var tokenMappings []model.User
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
			primitive.E{Key: "$or", Value: innerFilter},
		}

		err := sa.db.users.Find(filter, &tokenMappings, nil)
		if err != nil {
			return nil, err
		}

		recipients := []model.Recipient{}
		for _, user := range tokenMappings {
			recipients = append(recipients, model.Recipient{
				UserID:               user.UserID,
				NotificationDisabled: user.NotificationsDisabled,
			})
		}

		return recipients, nil
	}
	return nil, fmt.Errorf("no mapped recipients for the input criterias")
}

// UpdateUserByID Updates users notification enabled flag
func (sa Adapter) UpdateUserByID(userID string, notificationsDisabled bool) (*model.User, error) {
	if userID != "" {
		filter := bson.D{primitive.E{Key: "user_id", Value: userID}}

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

		return sa.FindUserByID(userID)
	}
	return nil, nil
}

// SubscribeToTopic subscribes the token to a topic
func (sa Adapter) SubscribeToTopic(token string, userID *string, topic string) error {
	var err error
	if userID != nil {
		record, err := sa.FindUserByID(*userID)
		if err == nil && record != nil {
			if err == nil && record != nil && !record.HasTopic(topic) {
				filter := bson.D{primitive.E{Key: "user_id", Value: record.UserID}}
				update := bson.D{
					primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "date_updated", Value: time.Now().UTC()},
					}},
					primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "topics", Value: topic}}},
				}
				_, err = sa.db.users.UpdateOne(filter, update, nil)
				if err == nil {
					var topicRecord *model.Topic
					topicRecord, _ = sa.GetTopicByName(topic)
					if topicRecord == nil {
						sa.InsertTopic(&model.Topic{Name: topic}) // just try to append within the topics collection
					}
				}
			}
		}
	} else {
		return fmt.Errorf("user id is nil")
	}

	return err
}

// UnsubscribeToTopic unsubscribes the token from a topic
func (sa Adapter) UnsubscribeToTopic(token string, userID *string, topic string) error {
	var err error
	if userID != nil {
		record, err := sa.FindUserByID(*userID)
		if err == nil && record != nil {
			if err == nil && record != nil && record.HasTopic(topic) {
				filter := bson.D{primitive.E{Key: "user_id", Value: record.UserID}}
				update := bson.D{
					primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "date_updated", Value: time.Now().UTC()},
					}},
					primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "topics", Value: topic}}},
				}
				_, err = sa.db.users.UpdateOne(filter, update, nil)
				if err == nil {
					var topicRecord *model.Topic
					topicRecord, _ = sa.GetTopicByName(topic)
					if topicRecord == nil {
						sa.InsertTopic(&model.Topic{Name: topic}) // just try to append within the topics collection in case it's missing
					}
				}
			}
		}
	} else {
		return fmt.Errorf("user id is nil")
	}

	return err
}

// GetTopics gets all topics
func (sa Adapter) GetTopics() ([]model.Topic, error) {
	filter := bson.D{}
	var result []model.Topic

	err := sa.db.topics.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetTopicByName appends a new topic within the topics collection
func (sa Adapter) GetTopicByName(name string) (*model.Topic, error) {
	if name != "" {
		filter := bson.D{primitive.E{Key: "_id", Value: name}}
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
	filter := bson.D{primitive.E{Key: "_id", Value: topic.Name}}

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

// GetMessages Gets all messages according to the filter
func (sa Adapter) GetMessages(userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	filter := bson.D{}
	innerFilter := []interface{}{}
	if userID != nil {
		innerFilter = append(innerFilter, bson.D{primitive.E{Key: "user_id", Value: userID}})
	}
	if len(innerFilter) > 0 {
		filter = append(filter, primitive.E{Key: "recipients", Value: bson.D{primitive.E{Key: "$elemMatch", Value: bson.D{primitive.E{Key: "$or", Value: innerFilter}}}}})
	}
	if filterTopic != nil {
		filter = append(filter, primitive.E{Key: "topic", Value: filterTopic})
	}
	if len(messageIDs) > 0 {
		filter = append(filter, primitive.E{Key: "_id", Value: bson.M{"$in": messageIDs}})
	}
	if startDateEpoch != nil {
		seconds := *startDateEpoch / 1000
		timeValue := time.Unix(seconds, 0)
		filter = append(filter, primitive.E{Key: "date_created", Value: bson.D{primitive.E{Key: "$gte", Value: &timeValue}}})
	}
	if endDateEpoch != nil {
		seconds := *endDateEpoch / 1000
		timeValue := time.Unix(seconds, 0)
		filter = append(filter, primitive.E{Key: "date_created", Value: bson.D{primitive.E{Key: "$lte", Value: &timeValue}}})
	}

	findOptions := options.Find()
	if order != nil && *order == "asc" {
		findOptions.SetSort(bson.D{{"date_created", 1}})
	} else {
		findOptions.SetSort(bson.D{{"date_created", -1}})
	}
	if limit != nil {
		findOptions.SetLimit(*limit)
	}
	if offset != nil {
		findOptions.SetSkip(*offset)
	}

	var list []model.Message
	err := sa.db.messages.Find(filter, &list, findOptions)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// GetMessage gets a message by id
func (sa Adapter) GetMessage(ID string) (*model.Message, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: ID}}

	var message *model.Message
	err := sa.db.messages.FindOne(filter, &message, nil)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// CreateMessage creates a new message.
func (sa Adapter) CreateMessage(message *model.Message) (*model.Message, error) {
	if message.ID == nil {
		id := uuid.New().String()
		message.ID = &id
	}
	now := time.Now().UTC()
	message.DateUpdated = &now
	message.DateCreated = &now

	_, err := sa.db.messages.InsertOne(&message)
	if err != nil {
		fmt.Printf("warning: error while store message (%s) - %s", *message.ID, err)
		return nil, err
	}

	return message, nil
}

// UpdateMessage updates a message
func (sa Adapter) UpdateMessage(message *model.Message) (*model.Message, error) {
	if message != nil && message.ID != nil {
		persistedMessage, err := sa.GetMessage(*message.ID)
		if err != nil || persistedMessage == nil {
			return nil, fmt.Errorf("Message with id (%s) not found: %w", *message.ID, err)
		}

		filter := bson.D{primitive.E{Key: "_id", Value: message.ID}}

		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "priority", Value: message.Priority},
				primitive.E{Key: "recipients", Value: message.Recipients},
				primitive.E{Key: "topic", Value: message.Topic},
				primitive.E{Key: "subject", Value: message.Subject},
				primitive.E{Key: "body", Value: message.Body},
				primitive.E{Key: "date_updated", Value: time.Now().UTC()},
			}},
		}

		_, err = sa.db.messages.UpdateOne(filter, update, nil)
		if err != nil {
			fmt.Printf("warning: error while update message (%s) - %s", *message.ID, err)
			return nil, err
		}
	}

	return message, nil
}

// DeleteUserMessage removes the desired user from the recipients list
func (sa Adapter) DeleteUserMessage(userID string, messageID string) error {
	persistedMessage, err := sa.GetMessage(messageID)
	if err != nil || persistedMessage == nil {
		return fmt.Errorf("message with id (%s) not found: %s", messageID, err)
	}

	updatesRecipients := []model.Recipient{}
	for _, recipient := range persistedMessage.Recipients {
		if userID != *recipient.UserID {
			updatesRecipients = append(updatesRecipients, recipient)
		}
	}

	if len(updatesRecipients) != len(persistedMessage.Recipients) {
		filter := bson.D{primitive.E{Key: "_id", Value: messageID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "recipients", Value: updatesRecipients},
			}},
		}

		_, err = sa.db.messages.UpdateOne(filter, update, nil)
		if err != nil {
			fmt.Printf("warning: error while delete message (%s) for user (%s) %s", messageID, userID, err)
			return err
		}
	}

	return nil
}

// DeleteMessage deletes a message by id
func (sa Adapter) DeleteMessage(ID string) error {
	persistedMessage, err := sa.GetMessage(ID)
	if err != nil || persistedMessage == nil {
		return fmt.Errorf("message with id (%s) not found: %s", ID, err)
	}

	filter := bson.D{primitive.E{Key: "_id", Value: ID}}
	_, err = sa.db.messages.DeleteOne(filter, nil)
	if err != nil {
		fmt.Printf("warning: error while delete message (%s) - %s", ID, err)
		return err
	}

	return nil
}

// GetAllAppVersions gets all registered versions
func (sa Adapter) GetAllAppVersions() ([]model.AppVersion, error) {
	filter := bson.D{}

	var versions []model.AppVersion
	err := sa.db.appVersions.Find(filter, &versions, nil)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

// GetAllAppPlatforms gets all registered platforms
func (sa Adapter) GetAllAppPlatforms() ([]model.AppPlatform, error) {
	filter := bson.D{}

	var platforms []model.AppPlatform
	err := sa.db.appPlatforms.Find(filter, &platforms, nil)
	if err != nil {
		return nil, err
	}

	return platforms, nil
}

func abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error on aborting a transaction - %s", err)
	}
}
