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

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// BBsAPIsHandler handles the rest BBs APIs implementation
type BBsAPIsHandler struct {
	app *core.Application
}

// NewBBsAPIsHandler creates new rest Handler instance
func NewBBsAPIsHandler(app *core.Application) BBsAPIsHandler {
	return BBsAPIsHandler{app: app}
}

// sendMessageRequestBody message request body
type bbsSendMessageRequestBody struct {
	Async   *bool                      `json:"async"`
	Message Def.SharedReqCreateMessage `json:"message"`
} // @name sendMessageRequestBody

// SendMessage Sends a message to a user, list of users or a topic
// @Description Sends a message to a user, list of users or a topic
// @Tags BBs
// @ID BBsSendMessage
// @Param data body sendMessageRequestBody true "body json"
// @Produce plain
// @Success 200 {object} model.Message
// @Security BBsAuth
// @Router /bbs/message [post]
func (h BBsAPIsHandler) SendMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var bodyData bbsSendMessageRequestBody
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	inputMessage := bodyData.Message
	async := false //by default
	if bodyData.Async != nil {
		async = *bodyData.Async
	}

	if len(inputMessage.OrgId) == 0 || len(inputMessage.AppId) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "org or app id", nil, nil, http.StatusBadRequest, false)
	}

	if !claims.AppOrg().CanAccessAppOrg(inputMessage.AppId, inputMessage.OrgId) {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "org or app id", nil, nil, http.StatusForbidden, false)
	}

	orgID := inputMessage.OrgId
	appID := inputMessage.AppId

	time, priority, subject, body, inputData, inputRecipients, recipientsCriteria, recipientsAccountCriteria, topic := getMessageData(inputMessage)

	sender := model.Sender{Type: "system", User: &model.CoreAccountRef{UserID: claims.Subject, Name: claims.Name}}

	message, err := h.app.BBs.BBsCreateMessage(orgID, appID,
		sender, time, priority, subject, body, inputData, inputRecipients, recipientsCriteria,
		recipientsAccountCriteria, topic, async)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "message", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// DeleteMessage deletes a message
func (h BBsAPIsHandler) DeleteMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.BBs.BBsDeleteMessage(l, claims.Subject, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "message", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// sendMailRequestBody mail request body
type bbsSendMailRequestBody struct {
	ToMail  string `json:"to_mail"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
} // @name sendMailRequestBody

// SendMail Sends an email
// @Description Sends an email
// @Tags BBs
// @ID BBsSendEmail
// @Param data body sendMailRequestBody true "body json"
// @Produce plain
// @Success 200
// @Security BBsAuth
// @Router /bbs/mail [post]
func (h BBsAPIsHandler) SendMail(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var mailRequest *bbsSendMailRequestBody
	err := json.NewDecoder(r.Body).Decode(&mailRequest)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	err = h.app.BBs.BBsSendMail(mailRequest.ToMail, mailRequest.Subject, mailRequest.Body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "email", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// AddRecipients add recipients to an existing message
// @Description add recipient
// @Tags BBs
// @ID BBsAddRecipients
// @Param data body addRecipientBody true "body json"
// @Produce plain
// @Success 200
// @Security BBsAuth
// @Router /bbs/recipients/{id} [put]
func (h BBsAPIsHandler) AddRecipients(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	messageID := params["id"]
	if len(messageID) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}
	read := getBoolQueryParam(r, "read")
	mute := getBoolQueryParam(r, "mute")

	err := h.app.BBs.BBsAddRecipients(l, messageID, claims.OrgID, claims.AppID, claims.Subject, mute, read)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "recipients", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()

}
