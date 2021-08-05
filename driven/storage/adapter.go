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
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// FindFirebaseToken finds firebase token
func (sa Adapter) FindFirebaseToken(token string) (*model.FirebaseTokenMapping, error) {
	filter := bson.D{}
	if len(token) > 0 {
		filter = bson.D{
			primitive.E{Key: "_id", Value: token},
		}
	}

	var result *model.FirebaseTokenMapping
	err := sa.db.tokens.FindOne(filter, &result, nil)
	if err != nil {
		log.Fatalf("warning: error while retriving token (%s) - %s", token, err)
	}

	return result, err
}

// StoreFirebaseToken stores firebase token and links it to the user
func (sa Adapter) StoreFirebaseToken(token string, user *model.User) error {
	_, err := sa.storeFirebaseToken(token, user)
	return err
}

func (sa Adapter) storeFirebaseToken(token string, user *model.User) (*model.FirebaseTokenMapping, error) {
	filter := bson.D{}
	if len(token) > 0 {
		filter = bson.D{
			primitive.E{Key: "_id", Value: token},
		}
	}

	var result *model.FirebaseTokenMapping
	err := sa.db.tokens.FindOne(filter, &result, nil)
	if err != nil {
		fmt.Printf("warning: error while retriving token (%s) - %s\n", token, err)
	}

	if result == nil {
		result, err = sa.createFirebaseToken(token, user)
	} else {
		result, err = sa.updateFirebaseToken(token, user)
	}

	return result, err
}

func (sa Adapter) createFirebaseToken(token string, user *model.User) (*model.FirebaseTokenMapping, error) {
	record := &model.FirebaseTokenMapping{
		Token: token,
	}

	if user != nil {
		record.Uin = user.Uin
		record.Email = user.Email
	} else {
		record.Uin = nil
		record.Email = nil
	}
	now := time.Now()
	record.DateCreated = now
	record.DateUpdated = now

	_, err := sa.db.tokens.InsertOne(record)
	if err != nil {
		fmt.Printf("warning: error while inserting token (%s) - %s\n", token, err)
	}

	return record, err
}

func (sa Adapter) updateFirebaseToken(token string, user *model.User) (*model.FirebaseTokenMapping, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: token}}

	record := &model.FirebaseTokenMapping{
		Token: token,
	}

	if user != nil {
		record.Uin = user.Uin
		record.Email = user.Email
	} else {
		record.Uin = nil
		record.Email = nil
	}
	now := time.Now()
	record.DateUpdated = now

	err := sa.db.tokens.ReplaceOne(filter, record, nil)
	if err != nil {
		fmt.Printf("warning: error while updating token (%s) - %s\n", token, err)
	}

	return record, err
}

// GetFirebaseTokensBy Gets all tokens mapped to the recipients input list
func (sa Adapter) GetFirebaseTokensBy(recipients []model.Recipient) ([]string, error) {
	if len(recipients) > 0 {
		innerFilter := []interface{}{}
		for _, recipient := range recipients {
			if recipient.Uin != nil {
				innerFilter = append(innerFilter, bson.D{primitive.E{Key: "uin", Value: recipient.Uin}})
			}
			if recipient.Email != nil {
				innerFilter = append(innerFilter, bson.D{primitive.E{Key: "email", Value: recipient.Email}})
			}
			if recipient.Phone != nil {
				innerFilter = append(innerFilter, bson.D{primitive.E{Key: "phone", Value: recipient.Phone}})
			}
		}

		filter := bson.D{
			primitive.E{Key: "$or", Value: innerFilter},
		}

		var tokenMapping []model.FirebaseTokenMapping
		err := sa.db.tokens.Find(filter, &tokenMapping, nil)
		if err != nil {
			return nil, err
		}

		tokens := make([]string, len(tokenMapping))
		for i, token := range tokenMapping {
			tokens[i] = token.Token
		}

		return tokens, nil
	}
	return nil, fmt.Errorf("empty recient information")
}

// SubscribeToTopic subscribes the token to a topic
func (sa Adapter) SubscribeToTopic(token string, user *model.User, topic string) error {
	record, err := sa.FindFirebaseToken(token)
	if err == nil {
		if record == nil {
			record, err = sa.storeFirebaseToken(token, user)
		}
		if err == nil && record != nil {
			record.DateUpdated = time.Now()
			record.AddTopic(topic)

			filter := bson.D{primitive.E{Key: "_id", Value: record.Token}}
			err = sa.db.tokens.ReplaceOne(filter, record, nil)
			if err != nil {
				log.Fatalf("warning: error while subscribe (%s) to topic (%s) - %s\n", token, topic, err)
			} else {
				_, _ = sa.AppendTopic(&model.Topic{Name: &topic}) // just try to append within the topics collection
			}
		}
	}

	return err
}

// UnsubscribeToTopic unsubscribes the token from a topic
func (sa Adapter) UnsubscribeToTopic(token string, user *model.User, topic string) error {
	record, err := sa.FindFirebaseToken(token)
	if err == nil {
		if record == nil {
			record, err = sa.storeFirebaseToken(token, user)
		}
		if err == nil && record != nil {
			record.DateUpdated = time.Now()
			record.RemoveTopic(topic)

			filter := bson.D{primitive.E{Key: "_id", Value: record.Token}}
			err = sa.db.tokens.ReplaceOne(filter, record, nil)
			if err != nil {
				log.Fatalf("warning: error while subscribe (%s) to topic (%s) - %s\n", token, topic, err)
			}
		}
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
func (sa Adapter) GetMessages(uinFilter *string, emailFilter *string, phoneFilter *string, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	filter := bson.D{}
	innerFilter := []interface{}{}
	if uinFilter != nil {
		innerFilter = append(innerFilter, bson.D{primitive.E{Key: "uin", Value: uinFilter}})
	}
	if emailFilter != nil {
		innerFilter = append(innerFilter, bson.D{primitive.E{Key: "email", Value: emailFilter}})
	}
	if phoneFilter != nil {
		innerFilter = append(innerFilter, bson.D{primitive.E{Key: "phone", Value: phoneFilter}})
	}
	if len(innerFilter) > 0 {
		filter = append(filter, primitive.E{Key: "recipients", Value: bson.D{primitive.E{ Key: "$elemMatch", Value: bson.D{primitive.E{Key: "$or", Value: innerFilter}}}}})
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
