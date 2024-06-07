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
	"strings"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
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

	inputData := bodyData.Message

	if len(inputData.OrgId) == 0 || len(inputData.AppId) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "org or app id", nil, nil, http.StatusBadRequest, false)
	}

	if !claims.AppOrg().CanAccessAppOrg(inputData.AppId, inputData.OrgId) {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "org or app id", nil, nil, http.StatusForbidden, false)
	}

	orgID := inputData.OrgId
	appID := inputData.AppId
	sender := model.Sender{Type: "system", User: &model.CoreAccountRef{UserID: claims.Subject, Name: claims.Name}}

	inputMessage := getMessageData(inputData)
	inputMessage.OrgID = orgID
	inputMessage.AppID = appID
	inputMessage.Sender = sender

	inputMessages := []model.InputMessage{inputMessage} //only one message

	messages, err := h.app.BBs.BBsCreateMessages(inputMessages, false)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "message", nil, err, http.StatusInternalServerError, true)
	}
	if len(messages) == 0 {
		return l.HTTPResponseErrorData(logutils.MessageDataStatus(logutils.StatusError), "message", nil, nil, http.StatusInternalServerError, false)
	}

	data, err := json.Marshal(messages[0])
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// SendMessages sends messages
func (h BBsAPIsHandler) SendMessages(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

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

		if !claims.AppOrg().CanAccessAppOrg(m.AppId, m.OrgId) {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, "org or app id", nil, nil, http.StatusForbidden, false)
		}

		inputMessage := getMessageData(m)
		inputMessage.OrgID = m.OrgId
		inputMessage.AppID = m.AppId
		inputMessage.Sender = model.Sender{Type: "system", User: &model.CoreAccountRef{UserID: claims.Subject, Name: claims.Name}}

		inputMessages = append(inputMessages, inputMessage)
	}

	createdMessages, err := h.app.BBs.BBsCreateMessages(inputMessages, isBatch)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "message", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(createdMessages)
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

	messagesIDs := []string{id} // only one
	err := h.app.BBs.BBsDeleteMessages(l, claims.Subject, messagesIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "message", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// DeleteMessages deletes messages
func (h BBsAPIsHandler) DeleteMessages(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	idsParam := getStringQueryParam(r, "ids")
	if idsParam == nil {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("ids"), nil, http.StatusBadRequest, false)
	}
	ids := strings.Split(*idsParam, ",")
	if len(ids) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypePathParam, logutils.StringArgs("ids"), nil, http.StatusBadRequest, false)
	}

	err := h.app.BBs.BBsDeleteMessages(l, claims.Subject, ids)
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
func (h BBsAPIsHandler) AddRecipients(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	messageID := params["message-id"]
	if len(messageID) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var bodyData Def.BbsReqAddRecipients
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}
	if len(bodyData) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "no data", nil, nil, http.StatusBadRequest, false)
	}

	recipients := make([]model.InputMessageRecipient, len(bodyData))
	for i, item := range bodyData {
		if len(item.UserId) == 0 {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, "no user id data", nil, nil, http.StatusBadRequest, false)
		}
		recipients[i] = model.InputMessageRecipient{UserID: item.UserId, Mute: item.Mute}
	}

	recipientResult, err := h.app.BBs.BBsAddRecipients(l, claims.Subject, messageID, recipients)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "recipients", nil, err, http.StatusInternalServerError, true)
	}
	data, err := json.Marshal(recipientResult)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// DeleteRecipients delete recipients from an existing message
func (h BBsAPIsHandler) DeleteRecipients(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	messageID := params["message-id"]
	if len(messageID) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var bodyData Def.BbsReqRemoveRecipients
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}
	if len(bodyData.UsersIds) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "no data", nil, nil, http.StatusBadRequest, false)
	}
	usersIDs := bodyData.UsersIds

	err = h.app.BBs.BBsDeleteRecipients(l, claims.Subject, messageID, usersIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "recipients", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}
