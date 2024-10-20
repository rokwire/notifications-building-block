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
	"encoding/json"
	"net/http"
	"notifications/core"
	"notifications/core/model"
	Def "notifications/driver/web/docs/gen"
	"strconv"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// InternalApisHandler handles the rest Admin APIs implementation
type InternalApisHandler struct {
	app *core.Application
}

// NewInternalApisHandler creates new rest Handler instance
func NewInternalApisHandler(app *core.Application) InternalApisHandler {
	return InternalApisHandler{app: app}
}

// SendMessage Sends a message to a user, list of users or a topic
// @Description Sends a message to a user, list of users or a topic
// @Tags Internal
// @ID InternalSendMessage
// @Param data body model.Message true "body json"
// @Produce plain
// @Success 200 {object} model.Message
// @Security InternalAuth
// @Router /int/message [post]
// @Deprecated
func (h InternalApisHandler) SendMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var inputData Def.SharedReqCreateMessage
	err := json.NewDecoder(r.Body).Decode(&inputData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	orgID := inputData.OrgId
	appID := inputData.AppId

	inputMessage := getMessageData(inputData)
	inputMessage.OrgID = orgID
	inputMessage.AppID = appID

	return h.processSendMessage(l, inputMessage, r)
}

// SendMessages sends messages
func (h InternalApisHandler) SendMessages(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	isBatch := false //false by default
	isBatchParam := r.URL.Query().Get("isBatch")
	if isBatchParam != "" {
		isBatch, _ = strconv.ParseBool(isBatchParam)
	}

	var bodyData Def.SharedReqCreateMessages
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	if len(bodyData) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "no data", nil, nil, http.StatusBadRequest, false)
	}

	inputMessages := []model.InputMessage{}
	//loop through the messages, validate each message and prepare InputMessage obj for every message
	for _, m := range bodyData {
		if len(m.OrgId) == 0 || len(m.AppId) == 0 {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, "org or app id", nil, nil, http.StatusBadRequest, false)
		}

		inputMessage := getMessageData(m)
		inputMessage.OrgID = m.OrgId
		inputMessage.AppID = m.AppId
		inputMessage.Sender = model.Sender{Type: "system"}

		inputMessages = append(inputMessages, inputMessage)
	}

	createdMessages, err := h.app.Services.CreateMessages(inputMessages, isBatch)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "message", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(createdMessages)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// sendMessageRequestBody message request body
type sendMessageRequestBody struct {
	Async   *bool                      `json:"async"`
	Message Def.SharedReqCreateMessage `json:"message"`
} // @name sendMessageRequestBody

// SendMessageV2 Sends a message to a user, list of users or a topic
// @Description Sends a message to a user, list of users or a topic
// @Tags Internal
// @ID InternalSendMessageV2
// @Param data body sendMessageRequestBody true "body json"
// @Produce plain
// @Success 200 {object} model.Message
// @Security InternalAuth
// @Router /int/v2/message [post]
func (h InternalApisHandler) SendMessageV2(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var bodyData sendMessageRequestBody
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	inputData := bodyData.Message

	orgID := inputData.OrgId
	appID := inputData.AppId

	inputMessage := getMessageData(inputData)
	inputMessage.OrgID = orgID
	inputMessage.AppID = appID

	return h.processSendMessage(l, inputMessage, r)
}

func (h InternalApisHandler) processSendMessage(l *logs.Log,
	inputMessage model.InputMessage, r *http.Request) logs.HTTPResponse {

	if len(inputMessage.OrgID) == 0 || len(inputMessage.AppID) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "org or app id", nil, nil, http.StatusBadRequest, false)
	}

	sender := model.Sender{Type: "system"}
	inputMessage.Sender = sender

	message, err := h.app.Services.CreateMessage(inputMessage)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "message", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// sendMailRequestBody mail request body
type sendMailRequestBody struct {
	ToMail  string `json:"to_mail"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
} // @name sendMailRequestBody

// SendMail Sends an email
// @Description Sends an email
// @Tags Internal
// @ID InternalSendMail
// @Param data body sendMailRequestBody true "body json"
// @Produce plain
// @Success 200
// @Security InternalAuth
// @Router /int/mail [post]
func (h InternalApisHandler) SendMail(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var mailRequest *sendMailRequestBody
	err := json.NewDecoder(r.Body).Decode(&mailRequest)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	err = h.app.Services.SendMail(mailRequest.ToMail, mailRequest.Subject, mailRequest.Body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "email", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}
