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
	"notifications/driven/storage"
	"time"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string
	StoreToken(orgID string, appID string, tokenInfo *model.TokenInfo, userID string) error
	SubscribeToTopic(orgID string, appID string, token string, userID string, anonymous bool, topic string) error
	UnsubscribeToTopic(orgID string, appID string, token string, userID string, anonymous bool, topic string) error
	GetTopics(orgID string, appID string) ([]model.Topic, error)
	AppendTopic(*model.Topic) (*model.Topic, error)
	UpdateTopic(*model.Topic) (*model.Topic, error)
	FindUserByID(orgID string, appID string, userID string, l *logs.Log) (*model.User, error)
	UpdateUserByID(orgID string, appID string, userID string, notificationsEnabled bool) (*model.User, error)
	DeleteUserWithID(orgID string, appID string, userID string) error

	GetMessagesRecipientsDeep(orgID string, appID string, userID *string, read *bool, mute *bool, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.MessageRecipient, error)

	GetMessagesStats(orgID string, appID string, userID string) (*model.MessagesStats, error)
	GetMessage(orgID string, appID string, ID string) (*model.Message, error)
	GetUserMessage(orgID string, appID string, ID string, accountID string) (*model.Message, error)
	CreateMessage(inputMessage model.InputMessage) (*model.Message, error)
	CreateMessages(inputMessages []model.InputMessage, isBatch bool) ([]model.Message, error)
	UpdateMessage(userID *string, message *model.Message) (*model.Message, error)
	DeleteUserMessage(orgID string, appID string, userID string, messageID string) error
	DeleteMessage(orgID string, appID string, ID string) error
	UpdateReadMessage(orgID string, appID string, ID string, userID string) (*model.Message, error)
	UpdateAllUserMessagesRead(orgID string, appID string, userID string, read bool) error

	GetAllAppVersions(orgID string, appID string) ([]model.AppVersion, error)
	GetAllAppPlatforms(orgID string, appID string) ([]model.AppPlatform, error)

	SendMail(toEmail string, subject string, body string) error

	PushSubscription(orgID string, appID string) error

	GetConfig(id string, claims *tokenauth.Claims) (*model.Configs, error)
	GetConfigs(configType *string, claims *tokenauth.Claims) ([]model.Configs, error)
	CreateConfig(config model.Configs, claims *tokenauth.Claims) (*model.Configs, error)
	UpdateConfig(config model.Configs, claims *tokenauth.Claims) error
	DeleteConfig(id string, claims *tokenauth.Claims) error
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) StoreToken(orgID string, appID string, tokenInfo *model.TokenInfo, userID string) error {
	return s.app.storeToken(orgID, appID, tokenInfo, userID)
}

func (s *servicesImpl) SubscribeToTopic(orgID string, appID string, token string, userID string, anonymous bool, topic string) error {
	return s.app.subscribeToTopic(orgID, appID, token, userID, anonymous, topic)
}

func (s *servicesImpl) UnsubscribeToTopic(orgID string, appID string, token string, userID string, anonymous bool, topic string) error {
	return s.app.unsubscribeToTopic(orgID, appID, token, userID, anonymous, topic)
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

func (s *servicesImpl) GetMessagesRecipientsDeep(orgID string, appID string, userID *string, read *bool, mute *bool, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.MessageRecipient, error) {
	return s.app.getMessagesRecipientsDeep(orgID, appID, userID, read, mute, messageIDs, startDateEpoch, endDateEpoch, filterTopic, offset, limit, order)
}

func (s *servicesImpl) GetMessagesStats(orgID string, appID string, userID string) (*model.MessagesStats, error) {
	return s.app.getMessagesStats(orgID, appID, userID)
}

func (s *servicesImpl) GetMessage(orgID string, appID string, ID string) (*model.Message, error) {
	return s.app.getMessage(orgID, appID, ID)
}

func (s *servicesImpl) GetUserMessage(orgID string, appID string, ID string, accountID string) (*model.Message, error) {
	return s.app.getUserMessage(orgID, appID, ID, accountID)
}

func (s *servicesImpl) CreateMessage(inputMessage model.InputMessage) (*model.Message, error) {
	return s.app.createMessage(inputMessage)
}

func (s *servicesImpl) CreateMessages(inputMessages []model.InputMessage, isBatch bool) ([]model.Message, error) {
	return s.app.createMessages(inputMessages, isBatch)
}

func (s *servicesImpl) UpdateMessage(userID *string, message *model.Message) (*model.Message, error) {
	return s.app.updateMessage(userID, message)
}

func (s *servicesImpl) UpdateReadMessage(orgID string, appID string, ID string, userID string) (*model.Message, error) {
	return s.app.updateReadMessage(orgID, appID, ID, userID)
}

func (s *servicesImpl) UpdateAllUserMessagesRead(orgID string, appID string, userID string, read bool) error {
	return s.app.updateAllUserMessagesRead(orgID, appID, userID, read)
}

func (s *servicesImpl) DeleteUserMessage(orgID string, appID string, userID string, messageID string) error {
	return s.app.deleteUserMessage(orgID, appID, userID, messageID)
}

func (s *servicesImpl) DeleteMessage(orgID string, appID string, messageID string) error {
	return s.app.deleteMessage(orgID, appID, messageID)
}

func (s *servicesImpl) GetAllAppVersions(orgID string, appID string) ([]model.AppVersion, error) {
	return s.app.getAllAppVersions(orgID, appID)
}

func (s *servicesImpl) GetAllAppPlatforms(orgID string, appID string) ([]model.AppPlatform, error) {
	return s.app.getAllAppPlatforms(orgID, appID)
}

func (s *servicesImpl) FindUserByID(orgID string, appID string, userID string, l *logs.Log) (*model.User, error) {
	return s.app.findUserByID(orgID, appID, userID, l)
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

func (s *servicesImpl) PushSubscription(orgID string, appID string) error {
	return s.app.pushSubscription(orgID, appID)
}

func (s *servicesImpl) GetConfig(id string, claims *tokenauth.Claims) (*model.Configs, error) {
	return s.app.getConfig(id, claims)
}

func (s *servicesImpl) GetConfigs(configType *string, claims *tokenauth.Claims) ([]model.Configs, error) {
	return s.app.getConfigs(configType, claims)
}

func (s *servicesImpl) CreateConfig(config model.Configs, claims *tokenauth.Claims) (*model.Configs, error) {
	return s.app.createConfig(config, claims)
}

func (s *servicesImpl) UpdateConfig(config model.Configs, claims *tokenauth.Claims) error {
	return s.app.updateConfig(config, claims)
}

func (s *servicesImpl) DeleteConfig(id string, claims *tokenauth.Claims) error {
	return s.app.deleteConfig(id, claims)
}

// Admin exposes APIs for the driver adapters
type Admin interface {
	AdminGetMessagesStats(orgID string, appID string, adminAccountID string, source string, offset *int64, limit *int64, order *string) (map[int][]interface{}, error)
}

type adminImpl struct {
	app *Application
}

func (s *adminImpl) AdminGetMessagesStats(orgID string, appID string, adminAccountID string, source string, offset *int64, limit *int64, order *string) (map[int][]interface{}, error) {
	return s.app.adminGetMessagesStats(orgID, appID, adminAccountID, source, offset, limit, order)
}

// BBs exposes users related APIs used by the platform building blocks
type BBs interface {
	BBsCreateMessages(inputMessages []model.InputMessage, isBatch bool) ([]model.Message, error)
	BBsDeleteMessages(l *logs.Log, serviceAccountID string, messagesIDs []string) error
	BBsSendMail(toEmail string, subject string, body string) error
	BBsAddRecipients(l *logs.Log, serviceAccountID string, messageID string, recipients []model.InputMessageRecipient) ([]model.MessageRecipient, error)
	BBsDeleteRecipients(l *logs.Log, serviceAccountID string, messageID string, usersIDs []string) error
}

type bbsImpl struct {
	app *Application
}

func (s *bbsImpl) BBsCreateMessages(inputMessages []model.InputMessage, isBatch bool) ([]model.Message, error) {
	return s.app.bbsCreateMessages(inputMessages, isBatch)
}

func (s *bbsImpl) BBsDeleteMessages(l *logs.Log, serviceAccountID string, messagesIDs []string) error {
	return s.app.bbsDeleteMessages(l, serviceAccountID, messagesIDs)
}

func (s *bbsImpl) BBsSendMail(toEmail string, subject string, body string) error {
	return s.app.bbsSendMail(toEmail, subject, body)
}

func (s *bbsImpl) BBsAddRecipients(l *logs.Log, serviceAccountID string, messageID string, recipients []model.InputMessageRecipient) ([]model.MessageRecipient, error) {
	return s.app.bbsAddRecipients(l, serviceAccountID, messageID, recipients)
}

func (s *bbsImpl) BBsDeleteRecipients(l *logs.Log, serviceAccountID string, messageID string, usersIDs []string) error {
	return s.app.bbsDeleteRecipients(l, serviceAccountID, messageID, usersIDs)
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	RegisterStorageListener(storageListener storage.Listener)

	PerformTransaction(func(context storage.TransactionContext) error, int64) error

	LoadFirebaseConfigurations() ([]model.FirebaseConf, error)

	FindUsersByIDs(usersIDs []string) ([]model.User, error)
	FindUserByID(orgID string, appID string, userID string) (*model.User, error)
	InsertUser(orgID string, appID string, userID string) (*model.User, error)
	UpdateUserByID(orgID string, appID string, userID string, notificationsEnabled bool) (*model.User, error)
	DeleteUserWithID(orgID string, appID string, userID string) error

	FindUserByToken(orgID string, appID string, token string) (*model.User, error)
	StoreDeviceToken(orgID string, appID string, tokenInfo *model.TokenInfo, userID string) error
	GetDeviceTokensByRecipients(orgID string, appID string, recipient []model.MessageRecipient, criteriaList []model.RecipientCriteria) ([]string, error)
	GetUsersByTopicsWithContext(ctx context.Context, orgID string, appID string, topic []string) ([]model.User, error)
	GetUsersByRecipientCriteriasWithContext(ctx context.Context, orgID string, appID string, recipientCriterias []model.RecipientCriteria) ([]model.User, error)
	SubscribeToTopic(orgID string, appID string, token string, userID string, topic string) error
	UnsubscribeToTopic(orgID string, appID string, token string, userID string, topic string) error
	GetTopics(orgID string, appID string) ([]model.Topic, error)
	InsertTopic(*model.Topic) (*model.Topic, error)
	UpdateTopic(*model.Topic) (*model.Topic, error)

	FindMessagesRecipients(orgID string, appID string, messageID string, userID string) ([]model.MessageRecipient, error)
	FindMessagesRecipientsByMessageAndUsers(messageID string, usersIDs []string) ([]model.MessageRecipient, error)
	FindMessagesRecipientsByMessages(messagesIDs []string) ([]model.MessageRecipient, error)
	FindMessagesRecipientsDeep(orgID string, appID string, userID *string, read *bool, mute *bool, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.MessageRecipient, error)
	InsertMessagesRecipientsWithContext(ctx context.Context, items []model.MessageRecipient) error
	DeleteMessagesRecipientsForIDsWithContext(ctx context.Context, ids []string) error
	DeleteMessagesRecipientsForMessagesWithContext(ctx context.Context, messagesIDs []string) error

	FindMessagesWithContext(ctx context.Context, ids []string) ([]model.Message, error)
	FindMessagesByParams(orgID string, appID string, senderType string, senderAccountID *string, offset *int64, limit *int64, order *string) ([]model.Message, error)
	GetMessage(orgID string, appID string, ID string) (*model.Message, error)
	CreateMessageWithContext(ctx context.Context, message model.Message) (*model.Message, error)
	InsertMessagesWithContext(ctx context.Context, messages []model.Message) error
	UpdateMessage(message *model.Message) (*model.Message, error)
	DeleteUserMessageWithContext(ctx context.Context, orgID string, appID string, userID string, messageID string) error
	DeleteMessagesWithContext(ctx context.Context, ids []string) error
	GetMessagesStats(userID string) (*model.MessagesStats, error)
	UpdateUnreadMessage(ctx context.Context, orgID string, appID string, ID string, userID string) (*model.Message, error)
	UpdateAllUserMessagesRead(ctx context.Context, orgID string, appID string, userID string, read bool) error
	GetAllAppVersions(orgID string, appID string) ([]model.AppVersion, error)
	GetAllAppPlatforms(orgID string, appID string) ([]model.AppPlatform, error)

	InsertQueueDataItemsWithContext(ctx context.Context, items []model.QueueItem) error
	LoadQueueWithContext(ctx context.Context) (*model.Queue, error)
	SaveQueueWithContext(ctx context.Context, queue model.Queue) error
	SaveQueue(queue model.Queue) error

	FindQueueData(time *time.Time, limit int) ([]model.QueueItem, error)
	DeleteQueueData(ids []string) error
	DeleteQueueDataForMessagesWithContext(ctx context.Context, messagesIDs []string) error
	DeleteQueueDataForRecipientsWithContext(ctx context.Context, recipientsIDs []string) error

	FindConfig(configType string, appID string, orgID string) (*model.Configs, error)
	FindConfigByID(id string) (*model.Configs, error)
	FindConfigs(configType *string) ([]model.Configs, error)
	InsertConfig(config model.Configs) error
	UpdateConfig(config model.Configs) error
	DeleteConfig(id string) error
}

// Firebase is used to wrap all Firebase Messaging API functions
type Firebase interface {
	UpdateFirebaseConfigurations(firebaseConfs []model.FirebaseConf) error
	SendNotificationToToken(orgID string, appID string, token string, title string, body string, data map[string]string) error
	SendNotificationToTopic(orgID string, appID string, topic string, title string, body string, data map[string]string) error
	SubscribeToTopic(orgID string, appID string, token string, topic string) error
	UnsubscribeToTopic(orgID string, appID string, token string, topic string) error
}

// Mailer is used to wrap all Email Messaging functions
type Mailer interface {
	SendMail(toEmail string, subject string, body string) error
}

// Core exposes Core APIs for the driver adapters
type Core interface {
	RetrieveCoreUserAccountByCriteria(accountCriteria map[string]interface{}, appID *string, orgID *string) ([]model.CoreAccount, error)
}

// Airship is used to wrap all Airship Messaging API Functions
type Airship interface {
	SendNotificationToToken(orgID string, appID string, deviceToken string, title string, body string, data map[string]string) error
}
