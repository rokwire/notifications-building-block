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

package core

import "notifications/core/model"

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string
	StoreFirebaseToken(tokenInfo *model.TokenInfo, user *model.CoreToken) error
	SubscribeToTopic(token string, user *model.CoreToken, topic string) error
	UnsubscribeToTopic(token string, user *model.CoreToken, topic string) error
	GetTopics() ([]model.Topic, error)
	AppendTopic(*model.Topic) (*model.Topic, error)
	UpdateTopic(*model.Topic) (*model.Topic, error)

	GetMessages(userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error)
	GetMessage(ID string) (*model.Message, error)
	CreateMessage(user *model.CoreToken, message *model.Message) (*model.Message, error)
	UpdateMessage(user *model.CoreToken, message *model.Message) (*model.Message, error)
	DeleteUserMessage(user *model.CoreToken, messageID string) error
	DeleteMessage(ID string) error

	GetAllAppVersions() ([]model.AppVersion, error)
	GetAllAppPlatforms() ([]model.AppPlatform, error)
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) StoreFirebaseToken(tokenInfo *model.TokenInfo, user *model.CoreToken) error {
	return s.app.storeFirebaseToken(tokenInfo, user)
}

func (s *servicesImpl) SubscribeToTopic(token string, user *model.CoreToken, topic string) error {
	return s.app.subscribeToTopic(token, user, topic)
}

func (s *servicesImpl) UnsubscribeToTopic(token string, user *model.CoreToken, topic string) error {
	return s.app.unsubscribeToTopic(token, user, topic)
}

func (s *servicesImpl) GetTopics() ([]model.Topic, error) {
	return s.app.getTopics()
}

func (s *servicesImpl) AppendTopic(topic *model.Topic) (*model.Topic, error) {
	return s.app.appendTopic(topic)
}

func (s *servicesImpl) UpdateTopic(topic *model.Topic) (*model.Topic, error) {
	return s.app.updateTopic(topic)
}

func (s *servicesImpl) GetMessages(userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	return s.app.getMessages(userID, messageIDs, startDateEpoch, endDateEpoch, filterTopic, offset, limit, order)
}

func (s *servicesImpl) GetMessage(ID string) (*model.Message, error) {
	return s.app.getMessage(ID)
}

func (s *servicesImpl) CreateMessage(user *model.CoreToken, message *model.Message) (*model.Message, error) {
	return s.app.createMessage(user, message)
}

func (s *servicesImpl) UpdateMessage(user *model.CoreToken, message *model.Message) (*model.Message, error) {
	return s.app.updateMessage(user, message)
}

func (s *servicesImpl) DeleteUserMessage(user *model.CoreToken, messageID string) error {
	return s.app.deleteUserMessage(user, messageID)
}

func (s *servicesImpl) DeleteMessage(messageID string) error {
	return s.app.deleteMessage(messageID)
}

func (s *servicesImpl) GetAllAppVersions() ([]model.AppVersion, error) {
	return s.app.getAllAppVersions()
}

func (s *servicesImpl) GetAllAppPlatforms() ([]model.AppPlatform, error) {
	return s.app.getAllAppPlatforms()
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	FindUserByID(userID string) (*model.User, error)
	FindUserByToken(token string) (*model.User, error)
	StoreFirebaseToken(tokenInfo *model.TokenInfo, user *model.CoreToken) error
	GetFirebaseTokensByRecipients(recipient []model.Recipient) ([]string, error)
	GetRecipientsByTopic(topic string) ([]model.Recipient, error)
	GetRecipientsByRecipientCriterias(recipientCriterias []model.RecipientCriteria) ([]model.Recipient, error)
	SubscribeToTopic(token string, userID *string, topic string) error
	UnsubscribeToTopic(token string, userID *string, topic string) error
	GetTopics() ([]model.Topic, error)
	InsertTopic(*model.Topic) (*model.Topic, error)
	UpdateTopic(*model.Topic) (*model.Topic, error)

	GetMessages(userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error)
	GetMessage(ID string) (*model.Message, error)
	CreateMessage(message *model.Message) (*model.Message, error)
	UpdateMessage(message *model.Message) (*model.Message, error)
	DeleteUserMessage(userID string, messageID string) error
	DeleteMessage(ID string) error

	GetAllAppVersions() ([]model.AppVersion, error)
	GetAllAppPlatforms() ([]model.AppPlatform, error)
}

// Firebase is used to wrap all Firebase Messaging API functions
type Firebase interface {
	SendNotificationToToken(token string, title string, body string, data map[string]string) error
	SendNotificationToTopic(topic string, title string, body string, data map[string]string) error
	SubscribeToTopic(token string, topic string) error
	UnsubscribeToTopic(token string, topic string) error
}
