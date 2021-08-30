/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
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
			primitive.E{Key: "firebase_tokens", Value: token},
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
	}

	return result, err
}

// StoreFirebaseToken stores firebase token and links it to the user
func (sa Adapter) StoreFirebaseToken(token string, userID *string) error {
	err := sa.storeFirebaseToken(token, userID)
	return err
}

func (sa Adapter) storeFirebaseToken(token string, userID *string) error {

	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		tokenRecord, _ := sa.findUserByTokenWithContext(sessionContext, token)
		if tokenRecord == nil {
			if userID != nil {
				user, _ := sa.findUserByIDWithContext(sessionContext, *userID)
				if user != nil {
					err = sa.addTokenToUserWithContext(sessionContext, token, userID)
				} else {
					_, err = sa.createUserWithContext(sessionContext, token, userID)
				}
			}
		} else if tokenRecord.UserID != nil && tokenRecord.UserID != userID {
			err = sa.removeTokenFromUserWithContext(sessionContext, token, tokenRecord.UserID)
			if err != nil {
				fmt.Printf("error while unlinking token (%s) from user (%s)- %s\n", token, *tokenRecord.UserID, err)
				return err
			}
			err = sa.addTokenToUserWithContext(sessionContext, token, userID)
			if err != nil {
				fmt.Printf("error while linking token (%s) from user (%s)- %s\n", token, *userID, err)
				return err
			}
		}

		if err != nil {
			fmt.Printf("error while storing token (%s) to user (%s) %s\n", token, *userID, err)
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

	return err
}

func (sa Adapter) createUserWithContext(context context.Context, token string, userID *string) (*model.User, error) {

	now := time.Now()
	record := &model.User{
		ID:          uuid.NewString(),
		UserID:      userID,
		Tokens:      []string{token},
		Topics:      []string{},
		DateCreated: now,
		DateUpdated: now,
	}

	_, err := sa.db.users.InsertOneWithContext(context, &record)
	if err != nil {
		fmt.Printf("warning: error while inserting token (%s) - %s\n", token, err)
	}

	return record, err
}

func (sa Adapter) addTokenToUserWithContext(ctx context.Context, token string, userID *string) error {
	if userID != nil {
		// transaction
		err := sa.db.dbClient.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			err := sessionContext.StartTransaction()
			if err != nil {
				log.Printf("error starting a transaction - %s", err)
				return err
			}

			filter := bson.D{primitive.E{Key: "user_id", Value: userID}}

			update := bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "date_updated", Value: time.Now()},
				}},
				primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "firebase_tokens", Value: token}}},
			}

			_, err = sa.db.users.UpdateOneWithContext(sessionContext, filter, &update, nil)
			if err != nil {
				fmt.Printf("warning: error while adding token (%s) to user (%s) %s\n", token, *userID, err)
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
			return fmt.Errorf("error while adding token (%s) to user (%s) %s", token, *userID, err)
		}
	}
	return nil
}

func (sa Adapter) removeTokenFromUserWithContext(ctx context.Context, token string, userID *string) error {
	if userID != nil {
		// transaction
		err := sa.db.dbClient.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			err := sessionContext.StartTransaction()
			if err != nil {
				log.Printf("error starting a transaction - %s", err)
				return err
			}

			filter := bson.D{primitive.E{Key: "user_id", Value: userID}}

			update := bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "date_updated", Value: time.Now()},
				}},
				primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "firebase_tokens", Value: token}}},
			}

			_, err = sa.db.users.UpdateOne(filter, &update, nil)
			if err != nil {
				fmt.Printf("warning: error while removing token (%s) from user (%s) %s\n", token, *userID, err)
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
			return fmt.Errorf("error while adding token (%s) to user (%s) %s", token, *userID, err)
		}
	}
	return nil
}

// GetFirebaseTokensByRecipients Gets all users mapped to the recipients input list
func (sa Adapter) GetFirebaseTokensByRecipients(recipients []model.Recipient) ([]string, error) {
	if len(recipients) > 0 {
		innerFilter := []interface{}{}
		for _, recipient := range recipients {
			if recipient.UserID != nil {
				innerFilter = append(innerFilter, bson.D{primitive.E{Key: "user_id", Value: recipient.UserID}})
			}
		}

		filter := bson.D{
			primitive.E{Key: "$or", Value: innerFilter},
		}

		var tokenMappings []model.User
		err := sa.db.users.Find(filter, &tokenMappings, nil)
		if err != nil {
			return nil, err
		}

		tokens := []string{}
		for _, tokenMapping := range tokenMappings {
			for _, token := range tokenMapping.Tokens {
				tokens = append(tokens, token)
			}
		}

		return tokens, nil
	}
	return nil, fmt.Errorf("empty recient information")
}

// SubscribeToTopic subscribes the token to a topic
func (sa Adapter) SubscribeToTopic(token string, userID *string, topic string) error {
	var err error
	if userID != nil {
		record, err := sa.FindUserByID(*userID)
		if err == nil && record != nil {
			if err == nil && record != nil {
				record.DateUpdated = time.Now()
				record.AddTopic(topic)

				filter := bson.D{primitive.E{Key: "_id", Value: record.ID}}
				update := bson.D{
					primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "date_updated", Value: time.Now()},
					}},
					primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "topics", Value: topic}}},
				}
				_, err = sa.db.users.UpdateOne(filter, update, nil)
				if err == nil {
					var topicRecord *model.Topic
					topicRecord, err = sa.GetTopicByName(topic)
					if err == nil {
						if topicRecord == nil {
							_, err = sa.InsertTopic(&model.Topic{Name: topic, UserIDs: []string{*userID}}) // just try to append within the topics collection
						} else {
							err = sa.AddUserIDToTopic(*userID, topic)
						}
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
			if err == nil && record != nil {
				record.DateUpdated = time.Now()
				record.AddTopic(topic)

				filter := bson.D{primitive.E{Key: "_id", Value: record.ID}}
				update := bson.D{
					primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "date_updated", Value: time.Now()},
					}},
					primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "topics", Value: topic}}},
				}
				_, err = sa.db.users.UpdateOne(filter, update, nil)
				if err == nil {
					var topicRecord *model.Topic
					topicRecord, _ = sa.GetTopicByName(topic)
					if topicRecord == nil {
						_, err = sa.InsertTopic(&model.Topic{Name: topic, UserIDs: []string{}}) // just try to append within the topics collection
					} else {
						err = sa.RemoveUserIDFromTopic(*userID, topic)
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
		now := time.Now()
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

// AddUserIDToTopic removes a user to a topic
func (sa Adapter) AddUserIDToTopic(userID string, topic string) error {
	filter := bson.D{primitive.E{Key: "_id", Value: topic}}

	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "date_updated", Value: time.Now()},
		}},
		primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "user_ids", Value: userID}}},
	}
	_, err := sa.db.topics.UpdateOne(filter, &update, nil)
	if err != nil {
		fmt.Printf("warning: error while add user (%s) topic (%s) - %s\n", userID, topic, err)
	}
	return err
}

// RemoveUserIDFromTopic removes a user from a topic
func (sa Adapter) RemoveUserIDFromTopic(userID string, topic string) error {
	filter := bson.D{primitive.E{Key: "_id", Value: topic}}

	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "date_updated", Value: time.Now()},
		}},
		primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "user_ids", Value: userID}}},
	}
	_, err := sa.db.topics.UpdateOne(filter, &update, nil)
	if err != nil {
		fmt.Printf("warning: error while add user (%s) topic (%s) - %s\n", userID, topic, err)
	}
	return err
}

// UpdateTopic updates a topic (for now only description is updatable)
func (sa Adapter) UpdateTopic(topic *model.Topic) (*model.Topic, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: topic.Name}}

	now := time.Now()
	topic.DateUpdated = now

	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "description", Value: topic.Description},
			primitive.E{Key: "date_updated", Value: topic.DateUpdated},
			primitive.E{Key: "user_ids", Value: topic.UserIDs},
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
func (sa Adapter) GetMessages(userID *string, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
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
	now := time.Now()
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
				primitive.E{Key: "date_updated", Value: time.Now()},
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

func abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error on aborting a transaction - %s", err)
	}
}
