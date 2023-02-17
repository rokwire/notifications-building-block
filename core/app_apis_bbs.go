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
	"notifications/core/model"
	"time"
)

func (app *Application) bbsCreateMessage(orgID string, appID string,
	sender model.Sender, mTime time.Time, priority int, subject string, body string, data map[string]string,
	inputRecipients []model.MessageRecipient, recipientsCriteriaList []model.RecipientCriteria,
	recipientAccountCriteria map[string]interface{}, topic *string, async bool) (*model.Message, error) {

	return app.sharedCreateMessage(orgID, appID, sender, mTime, priority, subject, body, data,
		inputRecipients, recipientsCriteriaList, recipientAccountCriteria, topic, async)
}

func (app *Application) bbsDeleteMessage(serviceAccountID string, messageID string) error {
	return nil
}

func (app *Application) bbsSendMail(toEmail string, subject string, body string) error {
	return app.sharedSendMail(toEmail, subject, body)
}
