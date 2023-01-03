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
	"github.com/rokwire/logging-library-go/v2/logs"
)

type queueLogic struct {
	logger *logs.Logger

	storage  Storage
	firebase Firebase
}

func (q queueLogic) start() {
	q.logger.Info("queueLogic start")

	//TODO set timer
}

func (q queueLogic) onQueuePush() {
	q.logger.Info("queueLogic onQueuePush")

	//TODO
}

/* TODO
func (app *Application) sendMessage(allRecipients []model.MessageRecipient, message model.Message, async bool) error {
	if len(allRecipients) == 0 {
		fmt.Print("no recipients")
		return nil
	}

	//send notifications only for mute=false
	recipients := []model.MessageRecipient{}
	for _, item := range allRecipients {
		if item.Mute == false {
			recipients = append(recipients, item)
		}
	}

	// retrieve tokens by recipients
	tokens, err := app.storage.GetFirebaseTokensByRecipients(
		message.OrgID, message.AppID, recipients, message.RecipientsCriteriaList)
	if err != nil {
		log.Printf("error on GetFirebaseTokensByRecipients: %s", err)
		return err
	}
	log.Printf("retrieve firebase tokens for message %s: %+v", message.ID, tokens)

	// send message to tokens
	if len(tokens) > 0 {
		if async {
			go app.sendNotifications(message, tokens)
		} else {
			app.sendNotifications(message, tokens)
		}
	}
	return nil
} */

/*
func (app *Application) sendNotifications(message model.Message, tokens []string) {
	for _, token := range tokens {
		sendErr := app.firebase.SendNotificationToToken(message.OrgID, message.AppID, token, message.Subject, message.Body, message.Data)
		if sendErr != nil {
			fmt.Printf("error send notification to token (%s): %s", token, sendErr)
		} else {
			log.Printf("message(%s:%s:%s) has been sent to token: %s", message.ID, message.Subject, message.Body, token)
		}
	}
} */
