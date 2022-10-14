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

package core

import (
	"context"
	"notifications/core/model"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string
	StoreFirebaseToken(orgID string, appID string, tokenInfo *model.TokenInfo, user *model.CoreToken) error
	SubscribeToTopic(orgID string, appID string, token string, user *model.CoreToken, topic string) error
	UnsubscribeToTopic(orgID string, appID string, token string, user *model.CoreToken, topic string) error
	GetTopics(orgID string, appID string) ([]model.Topic, error)
	AppendTopic(*model.Topic) (*model.Topic, error)
	UpdateTopic(*model.Topic) (*model.Topic, error)
	FindUserByID(orgID string, appID string, userID string) (*model.User, error)
	UpdateUserByID(orgID string, appID string, userID string, notificationsEnabled bool) (*model.User, error)
	DeleteUserWithID(orgID string, appID string, userID string) error

	GetMessages(orgID string, appID string, userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error)
	GetMessage(orgID string, appID string, ID string) (*model.Message, error)
	CreateMessage(user *model.CoreToken, message *model.Message, async bool) (*model.Message, error)
	UpdateMessage(user *model.CoreToken, message *model.Message) (*model.Message, error)
	DeleteUserMessage(user *model.CoreToken, messageID string) error
	DeleteMessage(ID string) error

	GetAllAppVersions() ([]model.AppVersion, error)
	GetAllAppPlatforms() ([]model.AppPlatform, error)

	SendMail(toEmail string, subject string, body string) error
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) StoreFirebaseToken(orgID string, appID string, tokenInfo *model.TokenInfo, user *model.CoreToken) error {
	return s.app.storeFirebaseToken(orgID, appID, tokenInfo, user)
}

func (s *servicesImpl) SubscribeToTopic(orgID string, appID string, token string, user *model.CoreToken, topic string) error {
	return s.app.subscribeToTopic(orgID, appID, token, user, topic)
}

func (s *servicesImpl) UnsubscribeToTopic(orgID string, appID string, token string, user *model.CoreToken, topic string) error {
	return s.app.unsubscribeToTopic(orgID, appID, token, user, topic)
}

func (s *servicesImpl) GetTopics(orgID string, appID string) ([]model.Topic, error) {
	return s.app.getTopics(orgID, appID)
}

func (s *servicesImpl) AppendTopic(topic *model.Topic) (*model.Topic, error) {
	return s.app.appendTopic(topic)
}

func (s *servicesImpl) UpdateTopic(topic *model.Topic) (*model.Topic, error) {
	return s.app.updateTopic(topic)
}

func (s *servicesImpl) GetMessages(orgID string, appID string, userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error) {
	return s.app.getMessages(orgID, appID, userID, messageIDs, startDateEpoch, endDateEpoch, filterTopic, offset, limit, order)
}

func (s *servicesImpl) GetMessage(orgID string, appID string, ID string) (*model.Message, error) {
	return s.app.getMessage(orgID, appID, ID)
}

func (s *servicesImpl) CreateMessage(user *model.CoreToken, message *model.Message, async bool) (*model.Message, error) {
	return s.app.createMessage(user, message, async)
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

func (s *servicesImpl) FindUserByID(orgID string, appID string, userID string) (*model.User, error) {
	return s.app.findUserByID(orgID, appID, userID)
}

func (s *servicesImpl) UpdateUserByID(orgID string, appID string, userID string, notificationsEnabled bool) (*model.User, error) {
	return s.app.updateUserByID(orgID, appID, userID, notificationsEnabled)
}

func (s *servicesImpl) DeleteUserWithID(orgID string, appID string, userID string) error {
	return s.app.deleteUserWithID(orgID, appID, userID)
}

func (s *servicesImpl) SendMail(toEmail string, subject string, body string) error {
	return s.app.sendMail(toEmail, subject, body)
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	FindUserByID(orgID string, appID string, userID string) (*model.User, error)
	UpdateUserByID(orgID string, appID string, userID string, notificationsEnabled bool) (*model.User, error)
	DeleteUserWithID(orgID string, appID string, userID string) error

	FindUserByToken(orgID string, appID string, token string) (*model.User, error)
	StoreFirebaseToken(orgID string, appID string, tokenInfo *model.TokenInfo, user *model.CoreToken) error
	GetFirebaseTokensByRecipients(orgID string, appID string, recipient []model.Recipient, criteriaList []model.RecipientCriteria) ([]string, error)
	GetRecipientsByTopic(orgID string, appID string, topic string) ([]model.Recipient, error)
	GetRecipientsByRecipientCriterias(orgID string, appID string, recipientCriterias []model.RecipientCriteria) ([]model.Recipient, error)
	SubscribeToTopic(orgID string, appID string, token string, userID *string, topic string) error
	UnsubscribeToTopic(orgID string, appID string, token string, userID *string, topic string) error
	GetTopics(orgID string, appID string) ([]model.Topic, error)
	InsertTopic(*model.Topic) (*model.Topic, error)
	UpdateTopic(*model.Topic) (*model.Topic, error)

	GetMessages(orgID string, appID string, userID *string, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.Message, error)
	GetMessage(orgID string, appID string, ID string) (*model.Message, error)
	CreateMessage(message *model.Message) (*model.Message, error)
	UpdateMessage(message *model.Message) (*model.Message, error)
	DeleteUserMessageWithContext(ctx context.Context, orgID string, appID string, userID string, messageID string) error
	DeleteMessageWithContext(ctx context.Context, orgID string, appID string, ID string) error

	GetAllAppVersions(orgID string, appID string) ([]model.AppVersion, error)
	GetAllAppPlatforms(orgID string, appID string) ([]model.AppPlatform, error)
}

// Firebase is used to wrap all Firebase Messaging API functions
type Firebase interface {
	SendNotificationToToken(orgID string, appID string, token string, title string, body string, data map[string]string) error
	SendNotificationToTopic(orgID string, appID string, topic string, title string, body string, data map[string]string) error
	SubscribeToTopic(orgID string, appID string, token string, topic string) error
	UnsubscribeToTopic(orgID string, appID string, token string, topic string) error
}

// Mailer is used to wrap all Email Messaging functions
type Mailer interface {
	SendMail(toEmail string, subject string, body string) error
}
