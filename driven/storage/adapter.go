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
	"strconv"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"

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
}

// Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
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

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS,
		multiTenancyOrgID: multiTenancyOrgID, multiTenancyAppID: multiTenancyAppID, logger: logger}
	return &Adapter{db: db}
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
		log.Printf("warning: error while retriving token (%s) - %s", token, err)
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

// StoreFirebaseToken stores firebase token
func (sa Adapter) StoreFirebaseToken(orgID string, appID string, tokenInfo *model.TokenInfo, userID string) error {
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			log.Printf("error starting a transaction - %s", err)
			return err
		}

		// Remove previous token no matter on with user is linked
		if tokenInfo.PreviousToken != nil {
			existingUser, _ := sa.findUserByTokenWithContext(sessionContext, orgID, appID, *tokenInfo.PreviousToken)
			if existingUser != nil {
				err = sa.removeTokenFromUserWithContext(sessionContext, orgID, appID, *tokenInfo.PreviousToken, existingUser.UserID)
				if err != nil {
					fmt.Printf("error while removing the previous token (%s) from user (%s)- %s\n", *tokenInfo.PreviousToken, userID, err)
					return err
				}
			}
		}

		userRecord, _ := sa.findUserByTokenWithContext(sessionContext, orgID, appID, *tokenInfo.Token)
		if userRecord == nil {
			existingUser, _ := sa.findUserByIDWithContext(sessionContext, orgID, appID, userID)
			if existingUser != nil {
				err = sa.addTokenToUserWithContext(sessionContext, orgID, appID, userID, *tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion)
			} else {
				_, err = sa.createUserWithContext(sessionContext, orgID, appID, userID, *tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion)
			}
		} else if userRecord.UserID != userID {
			err = sa.removeTokenFromUserWithContext(sessionContext, orgID, appID, *tokenInfo.Token, userRecord.UserID)
			if err != nil {
				fmt.Printf("error while unlinking token (%s) from user (%s)- %s\n", *tokenInfo.Token, userRecord.UserID, err)
				return err
			}

			existingUser, _ := sa.findUserByIDWithContext(sessionContext, orgID, appID, userID)
			if existingUser != nil {
				err = sa.addTokenToUserWithContext(sessionContext, orgID, appID, userID, *tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion)
			} else {
				_, err = sa.createUserWithContext(sessionContext, orgID, appID, userID, *tokenInfo.Token, tokenInfo.AppPlatform, tokenInfo.AppVersion)
			}
			if err != nil {
				fmt.Printf("error while linking token (%s) from user (%s)- %s\n", *tokenInfo.Token, userID, err)
				return err
			}
		}

		if err != nil {
			fmt.Printf("error while storing token (%s) to user (%s) %s\n", *tokenInfo.Token, userID, err)
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

func (sa Adapter) createUserWithContext(context context.Context, orgID string, appID string, userID string, token string, appPlatform *string, appVersion *string) (*model.User, error) {

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
		OrgID:          orgID,
		AppID:          appID,
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

func (sa Adapter) addTokenToUserWithContext(ctx context.Context, orgID string, appID string, userID string, token string, appPlatform *string, appVersion *string) error {
	// transaction
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "user_id", Value: userID},
	}

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

func (sa Adapter) removeTokenFromUserWithContext(ctx context.Context, orgID string, appID string, token string, userID string) error {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "user_id", Value: userID},
	}

	update := bson.D{
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

// GetFirebaseTokensByRecipients Gets all users mapped to the recipients input list
func (sa Adapter) GetFirebaseTokensByRecipients(orgID string, appID string, recipients []model.MessageRecipient, criteriaList []model.RecipientCriteria) ([]string, error) {
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
				for _, token := range tokenMapping.FirebaseTokens {
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

// GetUsersByTopicWithContext Gets all users for topic topic
func (sa Adapter) GetUsersByTopicWithContext(ctx context.Context, orgID string, appID string, topic string) ([]model.User, error) {
	if len(topic) > 0 {
		filter := bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "topics", Value: topic},
		}

		var tokenMappings []model.User
		err := sa.db.users.FindWithContext(ctx, filter, &tokenMappings, nil)
		if err != nil {
			return nil, err
		}

		result := []model.User{}
		for _, user := range tokenMappings {
			if user.HasTopic(topic) {
				result = append(result, user)
			}
		}

		return result, nil
	}
	return nil, fmt.Errorf("no mapped recipients to %s topic", topic)
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
	/*if userID != "" {

		err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
			err := sessionContext.StartTransaction()
			if err != nil {
				log.Printf("error starting a transaction - %s", err)
				abortTransaction(sessionContext)
				return err
			}

			messages, err := sa.GetMessages(orgID, appID, &userID, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			if err != nil {
				fmt.Printf("warning: unable to retrieve messages for user (%s): %s\n", userID, err)
				abortTransaction(sessionContext)
				return err
			}
			if len(messages) > 0 {
				for _, message := range messages {
					if message.Recipients != nil && len(message.Recipients) > 1 {
						err = sa.DeleteUserMessageWithContext(sessionContext, orgID, appID, userID, message.ID)
						if err != nil {
							fmt.Printf("warning: unable to unlink message(%s) for user(%s): %s\n", message.ID, userID, err)
						}
					} else {
						err = sa.DeleteMessageWithContext(sessionContext, orgID, appID, message.ID)
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
	*/
	return nil
}

// GetMessagesStats counts read/unread and muted/unmuted messages
func (sa *Adapter) GetMessagesStats(userID string, read bool, mute bool) (*model.MessagesStats, error) {
	pipeline := bson.A{
		bson.D{{"$match", bson.D{{"recipients.user_id", userID}}}},
		bson.D{
			{"$facet",
				bson.D{
					{"total_count",
						bson.A{
							bson.D{{"$count", "total_count"}},
						},
					},
					{"muted_count",
						bson.A{
							bson.D{{"$match", bson.D{{"recipients.mute", true}}}},
							bson.D{{"$count", "muted_count"}},
						},
					},
					{"not_muted_count",
						bson.A{
							bson.D{{"$match", bson.D{{"recipients.mute", bson.D{{"$ne", true}}}}}},
							bson.D{{"$count", "not_muted_count"}},
						},
					},
					{"read_count",
						bson.A{
							bson.D{{"$match", bson.D{{"recipients.read", true}}}},
							bson.D{{"$count", "read_count"}},
						},
					},
					{"not_read_count",
						bson.A{
							bson.D{{"$match", bson.D{{"recipients.read", bson.D{{"$ne", true}}}}}},
							bson.D{{"$count", "not_read_count"}},
						},
					},
				},
			},
		},
		bson.D{
			{"$project",
				bson.D{
					{"total_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$total_count.total_count",
									0,
								},
							},
						},
					},
					{"muted_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$muted_count.muted_count",
									0,
								},
							},
						},
					},
					{"not_muted_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$not_muted_count.not_muted_count",
									0,
								},
							},
						},
					},
					{"read_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$read_count.read_count",
									0,
								},
							},
						},
					},
					{"not_read_count",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$not_read_count.not_read_count",
									0,
								},
							},
						},
					},
				},
			},
		},
	}

	var stats []model.MessagesStats
	err := sa.db.messages.AggregateWithContext(nil, pipeline, &stats, nil)
	if err != nil {
		return nil, err
	}

	if len(stats) > 0 {
		stat := stats[0]
		return &stat, err
	}
	return nil, nil
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

// GetMessages Gets all messages according to the filter
func (sa Adapter) GetMessages(orgID string, appID string, userID *string, read *bool, mute *bool, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
	}
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

	if read != nil {
		if *read {
			filter = append(filter, primitive.E{Key: "recipients.read", Value: true})
		} else {
			// support of existing records where "read" field is missing
			filter = append(filter, primitive.E{Key: "recipients.read", Value: bson.M{"$ne": true}})
		}
	}

	if mute != nil {
		if *mute {
			filter = append(filter, primitive.E{Key: "recipients.mute", Value: true})
		} else {
			// support of existing records where "mute" field is missing
			filter = append(filter, primitive.E{Key: "recipients.mute", Value: bson.M{"$ne": true}})
		}
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
	/*if ctx == nil {
		ctx = context.Background()
	}
	persistedMessage, err := sa.GetMessage(orgID, appID, messageID)
	if err != nil || persistedMessage == nil {
		return fmt.Errorf("message with id (%s) not found: %s", messageID, err)
	}

	updatesRecipients := []model.MessageRecipient{}
	for _, recipient := range persistedMessage.Recipients {
		if userID != recipient.UserID {
			updatesRecipients = append(updatesRecipients, recipient)
		}
	}

	if len(updatesRecipients) != len(persistedMessage.Recipients) {
		filter := bson.D{
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "_id", Value: messageID}}
		update := bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "recipients", Value: updatesRecipients},
				primitive.E{Key: "date_updated", Value: time.Now().UTC()},
			}},
		}

		_, err = sa.db.messages.UpdateOneWithContext(ctx, filter, update, nil)
		if err != nil {
			fmt.Printf("warning: error while delete message (%s) for user (%s) %s", messageID, userID, err)
			return err
		}
	} */

	return nil
}

// DeleteMessageWithContext deletes a message by id
func (sa Adapter) DeleteMessageWithContext(ctx context.Context, orgID string, appID string, ID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	persistedMessage, err := sa.GetMessage(orgID, appID, ID)
	if err != nil || persistedMessage == nil {
		return fmt.Errorf("message with id (%s) not found: %s", ID, err)
	}

	filter := bson.D{
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "_id", Value: ID},
	}
	_, err = sa.db.messages.DeleteOneWithContext(ctx, filter, nil)
	if err != nil {
		fmt.Printf("warning: error while delete message (%s) - %s", ID, err)
		return err
	}

	return nil
}

// UpdateUnreadMessage updates a unread message in the recipients to read
func (sa Adapter) UpdateUnreadMessage(ctx context.Context, orgID string, appID string, ID string, userID string) (*model.Message, error) {
	read := true
	filter := bson.D{primitive.E{Key: "_id", Value: ID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "read", Value: read},
			primitive.E{Key: "date_updated", Value: time.Now().UTC()},
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
			primitive.E{Key: "date_updated", Value: time.Now().UTC()},
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
