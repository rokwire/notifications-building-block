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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"notifications/core/model"
	"strconv"
	"time"
)

//Adapter implements the Storage interface
type Adapter struct {
	db *database
}

//Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
	return err
}

type FirebaseTokenMapping struct {
	Token    string  `json:"firebase_token" bson:"_id"`
	DeviceID *string `json:"device_id" bson:"device_id"`
	Uin      *string `json:"uin" bson:"uin"`
	Email    *string `json:"email" bson:"email"`
	Phone    *string `json:"phone" bson:"phone"`
}

//NewStorageAdapter creates a new storage adapter instance
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

func (sa Adapter) StoreFirebaseToken(token string, user *model.User) error {
	filter := bson.D{}
	if len(token) > 0 {
		filter = bson.D{
			primitive.E{Key: "_id", Value: token},
		}
	}

	var result *FirebaseTokenMapping
	err := sa.db.tokens.FindOne(filter, &result, nil)
	if err != nil {
		fmt.Println("warning: error while retriving token (%s) - %s", token, err)
	}

	if result == nil {
		err = sa.createFirebaseToken(token, user)
	} else {
		err = sa.updateFirebaseToken(token, user)
	}

	return nil
}

func (sa Adapter) createFirebaseToken(token string, user *model.User) error {
	record := &FirebaseTokenMapping{
		Token: token,
	}

	if user != nil {
		record.Uin = user.Uin
		record.Email = user.Email
	} else {
		record.Uin = nil
		record.Email = nil
	}

	_, err := sa.db.tokens.InsertOne(record)
	if err != nil {
		fmt.Printf("warning: error while inserting token (%s) - %s\n", token, err)
	}

	return err
}

func (sa Adapter) updateFirebaseToken(token string, user *model.User) error {
	filter := bson.D{primitive.E{Key: "_id", Value: token}}

	record := &FirebaseTokenMapping{
		Token: token,
	}

	if user != nil {
		record.Uin = user.Uin
		record.Email = user.Email
	} else {
		record.Uin = nil
		record.Email = nil
	}

	err := sa.db.tokens.ReplaceOne(filter, record, nil)
	if err != nil {
		fmt.Printf("warning: error while updating token (%s) - %s\n", token, err)
	}

	return err
}

func (sa Adapter) GetFirebaseTokensBy(recipients []model.Recipient) ([]string, error) {
	if len(recipients) > 0 {
		innerFilter := []interface{}{}
		for _, recipient := range recipients{
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

		var tokenMapping []FirebaseTokenMapping
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
