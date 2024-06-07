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

package web

import (
	"fmt"
	"net/http"
	"notifications/core/model"
	Def "notifications/driver/web/docs/gen"
	"strconv"
	"time"
)

func getStringQueryParam(r *http.Request, paramName string) *string {
	params, ok := r.URL.Query()[paramName]
	if ok && len(params[0]) > 0 {
		value := params[0]
		return &value
	}
	return nil
}

func getInt64QueryParam(r *http.Request, paramName string) *int64 {
	params, ok := r.URL.Query()[paramName]
	if ok && len(params[0]) > 0 {
		val, err := strconv.ParseInt(params[0], 0, 64)
		if err == nil {
			return &val
		}
	}
	return nil
}

func getBoolQueryParam(r *http.Request, paramName string) *bool {
	readFromQuery, ok := r.URL.Query()[paramName]
	if ok && len(readFromQuery[0]) > 0 {
		val, err := strconv.ParseBool(readFromQuery[0])
		if err == nil {
			return &val
		}
	}
	return nil
}

func getMessageData(inputMessage Def.SharedReqCreateMessage) model.InputMessage {
	mTime := time.Now()
	if inputMessage.Time != nil {
		mTime = time.Unix(*inputMessage.Time, 0)
	}

	priority := inputMessage.Priority
	subject := inputMessage.Subject
	body := inputMessage.Body
	inputData := make(map[string]string, len(inputMessage.Data))
	for key, value := range inputMessage.Data {
		inputData[key] = fmt.Sprintf("%v", value)
	}

	//TODO add topic to topics array
	topics := append(inputMessage.Topics, *inputMessage.Topic)
	inputRecipients := messagesRecipientsListFromDef(inputMessage.Recipients)
	recipientsCriteria := recipientsCriteriaListFromDef(inputMessage.RecipientsCriteriaList)
	recipientsAccountCriteria := inputMessage.RecipientAccountCriteria

	return model.InputMessage{ID: inputMessage.Id, Time: mTime, Priority: priority, Subject: subject,
		Body: body, Data: inputData, Topic: inputMessage.Topic, Topics: topics, InputRecipients: inputRecipients,
		RecipientsCriteriaList: recipientsCriteria, RecipientAccountCriteria: recipientsAccountCriteria}
}
