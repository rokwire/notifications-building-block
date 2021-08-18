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
	StoreFirebaseToken(token string, userID *string) error
	SubscribeToTopic(token *string, userID *string, topic string) error
	UnsubscribeToTopic(token *string, userID *string, topic string) error
	GetTopics() ([]model.Topic, error)
	AppendTopic(*model.Topic) (*model.Topic, error)
	UpdateTopic(*model.Topic) (*model.Topic, error)

	SendMessage(user *model.ShibbolethUser, message *model.Message) (*model.Message, error)
	GetMessages(userID *string, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error)
	GetMessage(ID string) (*model.Message, error)
	CreateMessage(message *model.Message) (*model.Message, error)
	UpdateMessage(message *model.Message) (*model.Message, error)
	DeleteMessage(ID string) error
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) StoreFirebaseToken(token string, userID *string) error {
	return s.app.storeFirebaseToken(token, userID)
}

func (s *servicesImpl) SubscribeToTopic(token *string, userID *string, topic string) error {
	return s.app.subscribeToTopic(token, userID, topic)
}

func (s *servicesImpl) UnsubscribeToTopic(token *string, userID *string, topic string) error {
	return s.app.unsubscribeToTopic(token, userID, topic)
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

func (s *servicesImpl) SendMessage(user *model.ShibbolethUser, message *model.Message) (*model.Message, error) {
	return s.app.sendMessage(user, message)
}

func (s *servicesImpl) GetMessages(userID *string, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	return s.app.getMessages(userID, filterTopic, offset, limit, order)
}

func (s *servicesImpl) GetMessage(ID string) (*model.Message, error) {
	return s.app.getMessage(ID)
}

func (s *servicesImpl) CreateMessage(message *model.Message) (*model.Message, error) {
	return s.app.createMessage(message)
}

func (s *servicesImpl) UpdateMessage(message *model.Message) (*model.Message, error) {
	return s.app.updateMessage(message)
}

func (s *servicesImpl) DeleteMessage(ID string) error {
	return s.app.deleteMessage(ID)
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	FindUserByID(userID string) (*model.FirebaseTokenMapping, error)
	FindUserByToken(token string) (*model.FirebaseTokenMapping, error)
	StoreFirebaseToken(token string, userID *string) error
	GetFirebaseTokensBy(recipient []model.Recipient) ([]string, error)
	SubscribeToTopic(token string, userID *string, topic string) error
	UnsubscribeToTopic(token string, userID *string, topic string) error
	GetTopics() ([]model.Topic, error)
	AppendTopic(*model.Topic) (*model.Topic, error)
	UpdateTopic(*model.Topic) (*model.Topic, error)

	GetMessages(userID *string, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error)
	GetMessage(ID string) (*model.Message, error)
	CreateMessage(message *model.Message) (*model.Message, error)
	UpdateMessage(message *model.Message) (*model.Message, error)
	DeleteMessage(ID string) error
}

// Firebase is used to wrap all Firebase Messaging API functions
type Firebase interface {
	SendNotificationToToken(token string, title string, body string) error
	SendNotificationToTopic(topic string, title string, body string) error
	SubscribeToTopic(token string, topic string) error
	UnsubscribeToTopic(token string, topic string) error
}
