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
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	db       *mongo.Database
	dbClient *mongo.Client
	logger   *logs.Logger

	users              *collectionWrapper
	topics             *collectionWrapper
	messages           *collectionWrapper
	messagesRecipients *collectionWrapper

	appVersions  *collectionWrapper
	appPlatforms *collectionWrapper

	firebaseConfigurations *collectionWrapper

	listeners []Listener

	multiTenancyOrgID string
	multiTenancyAppID string
}

func (m *database) start() error {

	log.Println("database -> start")

	//connect to the database
	clientOptions := options.Client().ApplyURI(m.mongoDBAuth)
	connectContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	client, err := mongo.Connect(connectContext, clientOptions)
	cancel()
	if err != nil {
		return err
	}

	//ping the database
	pingContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	err = client.Ping(pingContext, nil)
	cancel()
	if err != nil {
		return err
	}

	//apply checks
	db := client.Database(m.mongoDBName)

	users := &collectionWrapper{database: m, coll: db.Collection("users")}
	err = m.applyUsersChecks(users)
	if err != nil {
		return err
	}

	topics := &collectionWrapper{database: m, coll: db.Collection("topics")}
	err = m.applyTopicsChecks(topics)
	if err != nil {
		return err
	}

	messages := &collectionWrapper{database: m, coll: db.Collection("messages")}
	err = m.applyMessagesChecks(messages)
	if err != nil {
		return err
	}

	messagesRecipients := &collectionWrapper{database: m, coll: db.Collection("messages_recipients")}
	err = m.applyMessagesRecipientsChecks(messagesRecipients)
	if err != nil {
		return err
	}

	appPlatforms := &collectionWrapper{database: m, coll: db.Collection("app_platforms")}
	err = m.applyPlatformsChecks(appPlatforms)
	if err != nil {
		return err
	}

	appVersions := &collectionWrapper{database: m, coll: db.Collection("app_versions")}
	err = m.applyVersionsChecks(appVersions)
	if err != nil {
		return err
	}

	firebaseConfigurations := &collectionWrapper{database: m, coll: db.Collection("firebase_configurations")}
	err = m.applyFirebaseConfigurationsChecks(firebaseConfigurations)
	if err != nil {
		return err
	}

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.users = users
	m.topics = topics
	m.messages = messages
	m.messagesRecipients = messagesRecipients
	m.appPlatforms = appPlatforms
	m.appVersions = appVersions
	m.firebaseConfigurations = firebaseConfigurations

	//apply multi-tenancy data manipulation
	err = m.fixMultiTenancyData(client, users, topics, messages, appVersions, appPlatforms)
	if err != nil {
		return err
	}

	go m.firebaseConfigurations.Watch(nil)

	return nil
}

// it adds org id and app id for the current data to match the multi-tenancy requirements
func (m *database) fixMultiTenancyData(client *mongo.Client, users *collectionWrapper, topics *collectionWrapper,
	messages *collectionWrapper, appVersions *collectionWrapper, appPlatforms *collectionWrapper) error {

	orgID := m.multiTenancyOrgID
	appID := m.multiTenancyAppID
	fn := func(sessionContext mongo.SessionContext) error {
		//start transaction
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a multi-tenancy data fix transaction - %s", err)
			return err
		}

		/// check if the data fix has been applied
		filter := bson.D{primitive.E{Key: "org_id", Value: orgID}, primitive.E{Key: "app_id", Value: appID}}
		// check users collection
		usersCount, err := users.CountDocumentsWithContext(sessionContext, filter)
		if err != nil {
			log.Println("error checking users count")
			return err
		}
		// check messages collection
		messagesCount, err := messages.CountDocumentsWithContext(sessionContext, filter)
		if err != nil {
			log.Println("error checking messages count")
			return err
		}

		//checking only one collection is enough as all this fixing data happens in a transaction.
		if usersCount == 0 && messagesCount == 0 {
			log.Printf("multi-tenancy data has NOT been applied, users:%d messages:%d - applying data fix..", usersCount, messagesCount)

			updatefilter := bson.D{}
			update := bson.D{
				primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "org_id", Value: orgID},
					primitive.E{Key: "app_id", Value: appID},
				}},
			}

			//users
			usersTimeout := time.Minute * time.Duration(2)
			_, err := users.UpdateManyWithContextTimeout(sessionContext, updatefilter, update, nil, usersTimeout) //long timeout
			if err != nil {
				log.Printf("error updating users - %s", err)
				return err
			}

			//topics
			_, err = topics.UpdateManyWithContext(sessionContext, updatefilter, update, nil)
			if err != nil {
				log.Printf("error updating topics - %s", err)
				return err
			}

			//messages
			messagesTimeout := time.Minute * time.Duration(2)
			_, err = messages.UpdateManyWithContextTimeout(sessionContext, updatefilter, update, nil, messagesTimeout) //long timeout
			if err != nil {
				log.Printf("error updating messages - %s", err)
				return err
			}

			//app versions
			_, err = appVersions.UpdateManyWithContext(sessionContext, updatefilter, update, nil)
			if err != nil {
				log.Printf("error updating app versions - %s", err)
				return err
			}

			//app platforms
			_, err = appPlatforms.UpdateManyWithContext(sessionContext, updatefilter, update, nil)
			if err != nil {
				log.Printf("error updating app platforms - %s", err)
				return err
			}
		} else {
			log.Println("multi-tenancy data has been applied, nothing to do")
			return nil
		}

		//commit the transaction
		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			abortTransaction(sessionContext)
			fmt.Printf("error on commiting multi-tenancy data fix transaction - %s", err)
			return err
		}
		log.Println("multi-tenancy data fix completed")
		return nil
	}
	err := client.UseSession(context.Background(), fn)
	return err
}

func (m *database) applyMessagesChecks(messages *collectionWrapper) error {
	log.Println("apply messages checks.....")

	//add compound unique index - org_id + app_id
	err := messages.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	indexes, _ := messages.ListIndexes()
	indexMapping := map[string]interface{}{}
	if indexes != nil {
		for _, index := range indexes {
			name := index["name"].(string)
			indexMapping[name] = index
		}
	}

	if indexMapping["recipients.user_id_1"] == nil {
		err := messages.AddIndex(
			bson.D{
				primitive.E{Key: "recipients.user_id", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["date_created_1"] == nil {
		err := messages.AddIndex(
			bson.D{
				primitive.E{Key: "date_created", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["date_updated_1"] == nil {
		err := messages.AddIndex(
			bson.D{
				primitive.E{Key: "date_updated", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["date_sent_1"] == nil {
		err := messages.AddIndex(
			bson.D{
				primitive.E{Key: "date_sent", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	log.Println("apply messages passed")
	return nil
}

func (m *database) applyMessagesRecipientsChecks(messagesRecipients *collectionWrapper) error {
	log.Println("apply messages recipients checks.....")

	//add org id index
	err := messagesRecipients.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	//add app id index
	err = messagesRecipients.AddIndex(bson.D{primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	//add user id index
	err = messagesRecipients.AddIndex(bson.D{primitive.E{Key: "user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	//add messages id index
	err = messagesRecipients.AddIndex(bson.D{primitive.E{Key: "message_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("apply messages recipients passed")
	return nil
}

func (m *database) applyUsersChecks(users *collectionWrapper) error {
	log.Println("apply users checks.....")

	//add compound unique index - org_id + app_id
	err := users.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	indexes, _ := users.ListIndexes()
	indexMapping := map[string]interface{}{}
	if indexes != nil {
		for _, index := range indexes {
			name := index["name"].(string)
			indexMapping[name] = index
		}
	}

	if indexMapping["user_id_1"] == nil {
		err := users.AddIndex(
			bson.D{
				primitive.E{Key: "user_id", Value: 1},
			}, true)
		if err != nil {
			return err
		}
	}

	if indexMapping["firebase_tokens.token_1"] == nil {
		err := users.AddIndex(
			bson.D{
				primitive.E{Key: "firebase_tokens.token", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["topics_1"] == nil {
		err := users.AddIndex(
			bson.D{
				primitive.E{Key: "topics", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	} else {
		err := users.DropIndex("topics_1") // rebuild nonunique index
		if err != nil {
			return err
		}

		err = users.AddIndex(
			bson.D{
				primitive.E{Key: "topics", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["date_created_1"] == nil {
		err := users.AddIndex(
			bson.D{
				primitive.E{Key: "date_created", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	if indexMapping["date_updated_1"] == nil {
		err := users.AddIndex(
			bson.D{
				primitive.E{Key: "date_updated", Value: 1},
			}, false)
		if err != nil {
			return err
		}
	}

	log.Println("apply users passed")
	return nil
}

func (m *database) applyTopicsChecks(topics *collectionWrapper) error {
	log.Println("apply topics checks.....")

	//add compound unique index - org_id + app_id
	err := topics.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("apply topics passed")
	return nil
}

func (m *database) applyVersionsChecks(appVersions *collectionWrapper) error {
	log.Println("apply app_versions checks.....")

	//add compound unique index - org_id + app_id
	err := appVersions.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	indexes, _ := appVersions.ListIndexes()
	indexMapping := map[string]interface{}{}
	if indexes != nil {
		for _, index := range indexes {
			name := index["name"].(string)
			indexMapping[name] = index
		}
	}

	if indexMapping["name"] == nil {
		err := appVersions.AddIndex(
			bson.D{
				primitive.E{Key: "name", Value: 1},
			}, true)
		if err != nil {
			return err
		}
	}

	log.Println("apply app_versions passed")
	return nil
}

func (m *database) applyPlatformsChecks(appPlatforms *collectionWrapper) error {
	log.Println("apply app_platforms checks.....")

	//add compound unique index - org_id + app_id
	err := appPlatforms.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	indexes, _ := appPlatforms.ListIndexes()
	indexMapping := map[string]interface{}{}
	if indexes != nil {
		for _, index := range indexes {
			name := index["name"].(string)
			indexMapping[name] = index
		}
	}

	if indexMapping["name"] == nil {
		err := appPlatforms.AddIndex(
			bson.D{
				primitive.E{Key: "name", Value: 1},
			}, true)
		if err != nil {
			return err
		}
	}

	log.Println("apply app_platforms passed")
	return nil
}

func (m *database) applyFirebaseConfigurationsChecks(fc *collectionWrapper) error {
	log.Println("apply firebase configurations checks.....")

	//add compound unique index - org_id + app_id
	err := fc.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}}, true)
	if err != nil {
		return err
	}

	log.Println("apply firebase configurations passed")
	return nil
}

func (m *database) onDataChanged(changeDoc map[string]interface{}) {
	if changeDoc == nil {
		return
	}
	log.Printf("onDataChanged: %+v\n", changeDoc)
	ns := changeDoc["ns"]
	if ns == nil {
		return
	}
	nsMap := ns.(map[string]interface{})
	coll := nsMap["coll"]

	switch coll {
	case "firebase_configurations":
		log.Println("firebase_configurations collection changed")

		for _, listener := range m.listeners {
			go listener.OnFirebaseConfigurationsUpdated()
		}
	}
}
