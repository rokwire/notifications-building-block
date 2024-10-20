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
	"errors"
	"fmt"
	"notifications/core/model"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/core-auth-library-go/v3/authutils"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) storeToken(orgID string, appID string, tokenInfo *model.TokenInfo, userID string) error {
	return app.storage.StoreDeviceToken(orgID, appID, tokenInfo, userID)
}

func (app *Application) subscribeToTopic(orgID string, appID string, token string, userID string, anonymous bool, topic string) error {
	var err error
	if !anonymous {
		err = app.storage.SubscribeToTopic(orgID, appID, token, userID, topic)
		if err == nil && token != "" {
			err = app.firebase.SubscribeToTopic(orgID, appID, token, topic)
		}
	} else if token != "" {
		// Treat this user as anonymous.
		err = app.firebase.SubscribeToTopic(orgID, appID, token, topic)
	}
	return err
}

func (app *Application) unsubscribeToTopic(orgID string, appID string, token string, userID string, anonymous bool, topic string) error {
	var err error
	if !anonymous {
		err = app.storage.UnsubscribeToTopic(orgID, appID, token, userID, topic)
		if err == nil && token != "" {
			err = app.firebase.UnsubscribeToTopic(orgID, appID, token, topic)
		}
	} else if token != "" {
		// Treat this user as anonymous.
		err = app.firebase.UnsubscribeToTopic(orgID, appID, token, topic)
	}
	return err
}

func (app *Application) getTopics(orgID string, appID string) ([]model.Topic, error) {
	return app.storage.GetTopics(orgID, appID)
}

func (app *Application) appendTopic(topic *model.Topic) (*model.Topic, error) {
	return app.storage.InsertTopic(topic)
}

func (app *Application) updateTopic(topic *model.Topic) (*model.Topic, error) {
	return app.storage.UpdateTopic(topic)
}

func (app *Application) createMessage(inputMessage model.InputMessage) (*model.Message, error) {
	inputMessages := []model.InputMessage{inputMessage} //only one
	messages, err := app.sharedCreateMessages(inputMessages, false)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, errors.New("error on creating message")
	}

	return &messages[0], nil //return only one
}

func (app *Application) createMessages(inputMessages []model.InputMessage, isBatch bool) ([]model.Message, error) {
	return app.sharedCreateMessages(inputMessages, isBatch)
}

func (app *Application) getMessagesRecipientsDeep(orgID string, appID string, userID *string, read *bool, mute *bool, messageIDs []string, startDateEpoch *int64, endDateEpoch *int64, filterTopic *string, offset *int64, limit *int64, order *string) ([]model.MessageRecipient, error) {
	return app.storage.FindMessagesRecipientsDeep(orgID, appID, userID, read, mute, messageIDs, startDateEpoch, endDateEpoch, filterTopic, offset, limit, order)
}

func (app *Application) getMessagesStats(orgID string, appID string, userID string) (*model.MessagesStats, error) {
	stats, _ := app.storage.GetMessagesStats(userID)
	return stats, nil
}

func (app *Application) getMessage(orgID string, appID string, ID string) (*model.Message, error) {
	return app.storage.GetMessage(orgID, appID, ID)
}

func (app *Application) getUserMessage(orgID string, appID string, ID string, accountID string) (*model.Message, error) {
	message, err := app.storage.GetMessage(orgID, appID, ID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		//no message for this id
		return nil, nil
	}

	//check if sender
	if message.IsSender(accountID) {
		return message, nil //it is sender
	}

	//check if recipient
	messagesRecipients, err := app.storage.FindMessagesRecipients(orgID, appID, ID, accountID)
	if err != nil {
		return nil, err
	}
	if len(messagesRecipients) > 0 {
		return message, err //it is recipient
	}

	return nil, nil //not sender, not recipient
}

func (app *Application) updateMessage(userID *string, message *model.Message) (*model.Message, error) {
	if message != nil {
		persistedMessage, err := app.storage.GetMessage(message.OrgID, message.AppID, message.ID)
		if err == nil && persistedMessage != nil {
			// If userID is nil, treat as system update, otherwise check sender match
			if userID == nil || (persistedMessage.Sender.User != nil && persistedMessage.Sender.User.UserID == *userID) {
				return app.storage.UpdateMessage(message)
			}
			return nil, fmt.Errorf("only creator can update the original message")
		}
	}
	return nil, fmt.Errorf("missing id or record")
}

func (app *Application) updateReadMessage(orgID string, appID string, ID string, userID string) (*model.Message, error) {
	updateReadMessage, _ := app.storage.UpdateUnreadMessage(context.Background(), orgID, appID, ID, userID)
	if updateReadMessage == nil {
		return nil, nil
	}
	return updateReadMessage, nil
}

func (app *Application) updateAllUserMessagesRead(orgID string, appID string, userID string, read bool) error {
	return app.storage.UpdateAllUserMessagesRead(context.Background(), orgID, appID, userID, read)
}

func (app *Application) deleteUserMessage(orgID string, appID string, userID string, messageID string) error {
	return app.storage.DeleteUserMessageWithContext(context.Background(), orgID, appID, userID, messageID)
}

func (app *Application) deleteMessage(orgID string, appID string, ID string) error {
	return app.storage.DeleteMessagesWithContext(context.Background(), []string{ID})
}

func (app *Application) getAllAppVersions(orgID string, appID string) ([]model.AppVersion, error) {
	return app.storage.GetAllAppVersions(orgID, appID)
}

func (app *Application) getAllAppPlatforms(orgID string, appID string) ([]model.AppPlatform, error) {
	return app.storage.GetAllAppPlatforms(orgID, appID)
}

func (app *Application) findUserByID(orgID string, appID string, userID string, l *logs.Log) (*model.User, error) {
	user, err := app.storage.FindUserByID(orgID, appID, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to find user(%s): %s", userID, err)
	}
	if user == nil {
		l.Infof("user not found for id {%s}, creating new user record", userID)
		user, err = app.storage.InsertUser(orgID, appID, userID)
		if err != nil {
			return nil, fmt.Errorf("unable to create user(%s): %s", userID, err)
		}
	}
	return user, nil
}

func (app *Application) updateUserByID(orgID string, appID string, userID string, notificationsDisabled bool) (*model.User, error) {
	return app.storage.UpdateUserByID(orgID, appID, userID, notificationsDisabled)
}

func (app *Application) deleteUserWithID(orgID string, appID string, userID string) error {
	user, err := app.storage.FindUserByID(orgID, appID, userID)
	if err != nil {
		return fmt.Errorf("unable to delete user(%s): %s", userID, err)
	}

	if user != nil {
		err = app.storage.DeleteUserWithID(orgID, appID, userID)
		if err != nil {
			return fmt.Errorf("unable to delete user(%s): %s", userID, err)
		}

		if user.Topics != nil && len(user.Topics) > 0 {
			for _, topic := range user.Topics {
				if user.DeviceTokens != nil && len(user.DeviceTokens) > 0 {
					for _, token := range user.DeviceTokens {
						err := app.firebase.UnsubscribeToTopic(orgID, appID, token.Token, topic)
						if err != nil {
							return fmt.Errorf("error unsubscribe user(%s) with token(%s) from topic(%s): %s", userID, token.Token, topic, err)
						}
					}
				}
			}
		}
	}

	return nil
}

func (app *Application) sendMail(toEmail string, subject string, body string) error {
	return app.sharedSendMail(toEmail, subject, body)
}

func (app *Application) pushSubscription(orgID string, appID string) error {
	return errors.New(logutils.Unimplemented)
}

func (app *Application) getConfig(id string, claims *tokenauth.Claims) (*model.Configs, error) {
	config, err := app.storage.FindConfigByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding config(%s): %s", id, err)

	}
	if config == nil {
		return nil, fmt.Errorf("error with config config(%s): %s", id, err)
	}

	err = claims.CanAccess(config.AppID, config.OrgID, config.System)
	if err != nil {
		return nil, fmt.Errorf("unable to access config: %s", err)
	}

	return config, nil
}

func (app *Application) getConfigs(configType *string, claims *tokenauth.Claims) ([]model.Configs, error) {
	configs, err := app.storage.FindConfigs(configType)
	if err != nil {
		return nil, fmt.Errorf("error finding configs(%s): %s", *configType, err)
	}

	allowedConfigs := make([]model.Configs, 0)
	for _, config := range configs {
		if err := claims.CanAccess(config.AppID, config.OrgID, config.System); err == nil {
			allowedConfigs = append(allowedConfigs, config)
		}
		allowedConfigs = append(allowedConfigs, config)
	}
	return allowedConfigs, nil
}

func (app *Application) createConfig(config model.Configs, claims *tokenauth.Claims) (*model.Configs, error) {
	// must be a system config if applying to all orgs
	if config.OrgID == authutils.AllOrgs && !config.System {
		return nil, fmt.Errorf("unauthorized to create config")

	}

	err := claims.CanAccess(config.AppID, config.OrgID, config.System)
	if err != nil {
		return nil, fmt.Errorf("unable to access config: %s", err)
	}

	config.ID = uuid.NewString()
	config.DateCreated = time.Now().UTC()
	err = app.storage.InsertConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to insert config")

	}
	return &config, nil
}

func (app *Application) updateConfig(config model.Configs, claims *tokenauth.Claims) error {
	// must be a system config if applying to all orgs
	if config.OrgID == authutils.AllOrgs && !config.System {
		return fmt.Errorf("unable to update config")
	}

	oldConfig, err := app.storage.FindConfig(config.Type, config.AppID, config.OrgID)
	if err != nil {
		return fmt.Errorf("unable  to update config")
	}
	if oldConfig == nil {
		return fmt.Errorf("unable to update config, old config is null")
	}

	//cannot update a system config if not a system admin
	if !claims.System && oldConfig.System {
		return fmt.Errorf("unable to update user, not s system admin")
	}
	err = claims.CanAccess(config.AppID, config.OrgID, config.System)
	if err != nil {
		return fmt.Errorf("unauthorized to update user")
	}

	now := time.Now().UTC()
	config.ID = oldConfig.ID
	config.DateUpdated = &now

	err = app.storage.UpdateConfig(config)
	if err != nil {
		return fmt.Errorf("unable to update user")
	}
	return nil
}

func (app *Application) deleteConfig(id string, claims *tokenauth.Claims) error {
	config, err := app.storage.FindConfigByID(id)
	if err != nil {
		return fmt.Errorf("unable to delete config")
	}
	if config == nil {
		return fmt.Errorf("unable to delete config, config is null")
	}

	err = claims.CanAccess(config.AppID, config.OrgID, config.System)
	if err != nil {
		return fmt.Errorf("unauthorized to delete config")
	}

	err = app.storage.DeleteConfig(id)
	if err != nil {
		return fmt.Errorf("unable to delete config")
	}
	return nil
}
