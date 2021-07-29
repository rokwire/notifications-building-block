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

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) storeFirebaseToken(token string, user *model.User) error {
	return app.storage.StoreFirebaseToken(token, user)
}

func (app *Application) sendMessage(message model.Message) error {
	if len(message.Recipients) > 0 {
		tokens, err := app.storage.GetFirebaseTokensBy(message.Recipients)
		if err != nil {
			return err
		}
		if len(tokens) > 0 {
			for _, token := range tokens {
				_ = app.firebase.SendNotificationToToken(token, message.Subject, message.Body)
			}
		}
	}
	return nil
}
