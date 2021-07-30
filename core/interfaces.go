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
	StoreFirebaseToken(token string, user *model.User) error
	SendMessage(message model.Message) error
	SubscribeToTopic(token string, user *model.User, topic string) error
	UnsubscribeToTopic(token string, user *model.User, topic string) error
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) StoreFirebaseToken(token string, user *model.User) error {
	return s.app.storeFirebaseToken(token, user)
}

func (s *servicesImpl) SendMessage(message model.Message) error {
	return s.app.sendMessage(message)
}

func (s *servicesImpl) SubscribeToTopic(token string, user *model.User, topic string) error {
	return s.app.subscribeToTopic(token, user, topic)
}

func (s *servicesImpl) UnsubscribeToTopic(token string, user *model.User, topic string) error {
	return s.app.unsubscribeToTopic(token, user, topic)
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	StoreFirebaseToken(token string, user *model.User) error
	GetFirebaseTokensBy(recipient []model.Recipient) ([]string, error)
	SubscribeToTopic(token string, user *model.User, topic string) error
	UnsubscribeToTopic(token string, user *model.User, topic string) error
}

type Firebase interface {
	SendNotificationToToken(token string, title string, body string) error
	SendNotificationToTopic(topic string, title string, body string) error
	SubscribeToTopic(token string, topic string) error
	UnsubscribeToTopic(token string, topic string) error
}
