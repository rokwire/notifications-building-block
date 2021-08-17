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
func (sa Adapter) FindUserByToken(token string) (*model.FirebaseTokenMapping, error) {
	return sa.findUserByTokenWithContext(context.Background(), token)
}

func (sa Adapter) findUserByTokenWithContext(context context.Context, token string) (*model.FirebaseTokenMapping, error) {
	filter := bson.D{}
	if len(token) > 0 {
		filter = bson.D{
			primitive.E{Key: "firebase_tokens", Value: token},
		}
	}

	var result *model.FirebaseTokenMapping
	err := sa.db.users.FindOneWithContext(context, filter, &result, nil)
	if err != nil {
		log.Printf("warning: error while retriving token (%s) - %s", token, err)
	}

	return result, err
}

// FindUserByID finds user by id
func (sa Adapter) FindUserByID(userID string) (*model.FirebaseTokenMapping, error) {
	return sa.findUserByIDWithContext(context.Background(), userID)
}

func (sa Adapter) findUserByIDWithContext(context context.Context, userID string) (*model.FirebaseTokenMapping, error) {
	filter := bson.D{}
	if len(userID) > 0 {
		filter = bson.D{
			primitive.E{Key: "user_id", Value: userID},
		}
	}

	var result *model.FirebaseTokenMapping
	err := sa.db.users.FindOneWithContext(context, filter, &result, nil)
	if err != nil {
		log.Printf("warning: error while retriving user (%s) - %s", userID, err)
	}

	return result, err
}

func (sa Adapter) FindUser(user string) (*model.FirebaseTokenMapping, error) {
	filter := bson.D{}
	if len(user) > 0 {
		filter = bson.D{
			primitive.E{Key: "user_id", Value: user},
		}
	}

	var result *model.FirebaseTokenMapping
	err := sa.db.users.FindOne(filter, &result, nil)
	if err != nil {
		log.Printf("warning: error while retriving user (%s) - %s", user, err)
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
				fmt.Printf("error while unlinking token (%s) from user (%s)- %s\n", token, tokenRecord.UserID, err)
				return err
			}
			err = sa.addTokenToUserWithContext(sessionContext, token, userID)
			if err != nil {
				fmt.Printf("error while linking token (%s) from user (%s)- %s\n", token, userID, err)
				return err
			}
		}

		if err != nil {
			fmt.Printf("error while storing token (%s) to user (%s) %s\n", token, userID, err)
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

func (sa Adapter) createUserWithContext(context context.Context, token string, userID *string) (*model.FirebaseTokenMapping, error) {
	record := &model.FirebaseTokenMapping{
		Tokens: []string{token},
		UserID: userID,
		ID:     uuid.NewString(),
	}

	now := time.Now()
	record.DateCreated = now
	record.DateUpdated = now

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
				fmt.Printf("warning: error while adding token (%s) to user (%s) %s\n", token, userID, err)
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
			return fmt.Errorf("error while adding token (%s) to user (%s) %s\n", token, userID, err)
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
				fmt.Printf("warning: error while removing token (%s) from user (%s) %s\n", token, userID, err)
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
			return fmt.Errorf("error while adding token (%s) to user (%s) %s\n", token, userID, err)
		}
	}
	return nil
}

/*func (sa Adapter) findToken(token string) (*model.FirebaseTokenMapping){
	filter := bson.D{}
	if len(token) > 0 {
		filter = bson.D{
			primitive.E{Key: "firebase_tokens", Value: token},
		}
	}

	var result *model.FirebaseTokenMapping
	err := sa.db.users.FindOne(filter, &result, nil)
	if err != nil {
		fmt.Printf("warning: error while retriving token (%s) - %s\n", token, err)
	}

	return result
}*/

// GetFirebaseTokensBy Gets all users mapped to the recipients input list
func (sa Adapter) GetFirebaseTokensBy(recipients []model.Recipient) ([]string, error) {
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

		var tokenMappings []model.FirebaseTokenMapping
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
	/*record, err := sa.FindFirebaseToken(token)
	if err != nil || record == nil {
		record, err = sa.storeFirebaseToken(token, user)
	}
	if err == nil && record != nil {
		if err == nil && record != nil {
			record.DateUpdated = time.Now()
			record.AddTopic(topic)

			filter := bson.D{primitive.E{Key: "_id", Value: record.Token}}
			err = sa.db.tokens.ReplaceOne(filter, record, nil)
			if err != nil {
				log.Printf("warning: error while subscribe (%s) to topic (%s) - %s\n", token, topic, err)
			} else {
				_, _ = sa.AppendTopic(&model.Topic{Name: &topic}) // just try to append within the topics collection
			}
		}
	}

	return err*/
	return nil
}

// UnsubscribeToTopic unsubscribes the token from a topic
func (sa Adapter) UnsubscribeToTopic(token string, userID *string, topic string) error {
	/*record, err := sa.FindFirebaseToken(token)
	if err != nil || record == nil {
		record, err = sa.storeFirebaseToken(token, user)
	}
	if err == nil && record != nil {
		if err == nil && record != nil {
			record.DateUpdated = time.Now()
			record.RemoveTopic(topic)

			filter := bson.D{primitive.E{Key: "_id", Value: record.Token}}
			err = sa.db.tokens.ReplaceOne(filter, record, nil)
			if err != nil {
				log.Printf("warning: error while subscribe (%s) to topic (%s) - %s\n", token, topic, err)
			}
		}
	}

	return err*/
	return nil
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

// AppendTopic appends a new topic within the topics collection
func (sa Adapter) AppendTopic(topic *model.Topic) (*model.Topic, error) {
	if topic.Name != nil {
		now := time.Now()
		topic.DateUpdated = &now
		topic.DateCreated = &now

		_, err := sa.db.topics.InsertOne(&topic)
		if err != nil {
			fmt.Printf("warning: error while store topic (%s) - %s\n", *topic.Name, err)
			return nil, err
		}
	}

	return topic, nil
}

// UpdateTopic updates a topic (for now only description is updatable)
func (sa Adapter) UpdateTopic(topic *model.Topic) (*model.Topic, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: topic.Name}}

	now := time.Now()
	topic.DateUpdated = &now

	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "description", Value: topic.Description},
			primitive.E{Key: "date_updated", Value: topic.DateUpdated},
		}},
	}

	_, err := sa.db.topics.UpdateOne(filter, &update, nil)
	if err != nil {
		fmt.Printf("warning: error while update topic (%s) - %s\n", *topic.Name, err)
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

	if message.Sent {
		message.DateSent = &now
	}

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
		if persistedMessage.Sent {
			return nil, fmt.Errorf("attempt to update already sent message")
		}
		if message.Sent && !persistedMessage.Sent {
			now := time.Now()
			message.DateSent = &now
		}

		filter := bson.D{primitive.E{Key: "_id", Value: message.ID}}

		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "priority", Value: message.Priority},
				primitive.E{Key: "recipients", Value: message.Recipients},
				primitive.E{Key: "topic", Value: message.Topic},
				primitive.E{Key: "subject", Value: message.Subject},
				primitive.E{Key: "body", Value: message.Body},
				primitive.E{Key: "sender", Value: message.Sender},
				primitive.E{Key: "date_updated", Value: time.Now()},
				primitive.E{Key: "date_sent", Value: message.DateSent},
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

// DeleteMessage deletes a message by id
func (sa Adapter) DeleteMessage(ID string) error {
	persistedMessage, err := sa.GetMessage(ID)
	if err != nil || persistedMessage == nil {
		return fmt.Errorf("message with id (%s) not found: %s", ID, err)
	}
	if persistedMessage.Sent {
		return fmt.Errorf("unable to delete message which is already sent")
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
