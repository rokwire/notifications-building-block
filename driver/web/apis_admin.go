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
	"errors"
	"net/http"
	"notifications/core"
	"notifications/core/model"

	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"

	Def "notifications/driver/web/docs/gen"

	"github.com/gorilla/mux"
)

// AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app *core.Application
}

// NewAdminApisHandler creates new rest Handler instance
func NewAdminApisHandler(app *core.Application) AdminApisHandler {
	return AdminApisHandler{app: app}
}

// GetTopics Gets all topics
// @Description Gets all topics
// @Tags Admin
// @ID AdminGetTopics
// @Success 200 {array} model.Topic
// @Security AdminUserAuth
// @Router /admin/topics [get]
func (h AdminApisHandler) GetTopics(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	topics, err := h.app.Services.GetTopics(claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "topics", nil, err, http.StatusBadRequest, true)
	}

	data, err := json.Marshal(topics)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// UpdateTopic Updated the topic
// @Description Updated the topic.
// @Tags Admin
// @ID UpdateTopic
// @Param data body model.Topic true "body json"
// @Success 200 {object} model.Topic
// @Security AdminUserAuth
// @Router /admin/topic [put]
func (h AdminApisHandler) UpdateTopic(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var topic *model.Topic
	err := json.NewDecoder(r.Body).Decode(&topic)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	topic.OrgID = claims.OrgID
	topic.AppID = claims.AppID

	_, err = h.app.Services.UpdateTopic(topic)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, "topic", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(topic)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// GetMessages Gets all messages. This api may be invoked with different filters in the query string
// @Description Gets all messages
// @Tags Admin
// @ID GetMessages
// @Param user query string false "user - filter by user"
// @Param topic query string false "topic - filter by topic"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Param start_date query string false "start_date - Start date filter in milliseconds as an integer epoch value"
// @Param end_date query string false "end_date - End date filter in milliseconds as an integer epoch value"
// @Success 200 {array} model.Message
// @Security AdminUserAuth
// @Router /admin/messages [get]
func (h AdminApisHandler) GetMessages(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	return l.HTTPResponseSuccess()

	/*userIDFilter := getStringQueryParam(r, "user")
	topicFilter := getStringQueryParam(r, "topic")
	offsetFilter := getInt64QueryParam(r, "offset")
	limitFilter := getInt64QueryParam(r, "limit")
	orderFilter := getStringQueryParam(r, "order")
	startDateFilter := getInt64QueryParam(r, "start_date")
	endDateFilter := getInt64QueryParam(r, "end_date")
	read := getBoolQueryParam(r, "read")
	mute := getBoolQueryParam(r, "mute")

	messages, err := h.app.Services.GetMessages(claims.OrgID, claims.AppID, userIDFilter, read, mute, nil, startDateFilter, endDateFilter, topicFilter, offsetFilter, limitFilter, orderFilter)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "messages", nil, err, http.StatusInternalServerError, true)
	}

	if messages == nil {
		messages = []model.Message{}
	}

	data, err := json.Marshal(messages)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data) */
}

// CreateMessage Creates a message
func (h AdminApisHandler) CreateMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var inputData Def.SharedReqCreateMessage
	err := json.NewDecoder(r.Body).Decode(&inputData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}
	if len(inputData.Body) == 0 {
		return l.HTTPResponseErrorAction(logutils.ActionGet, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	orgID := claims.OrgID
	appID := claims.AppID
	sender := model.Sender{Type: "administrative", User: &model.CoreAccountRef{UserID: claims.Subject, Name: claims.Name}}

	inputMessage := getMessageData(inputData)
	inputMessage.OrgID = orgID
	inputMessage.AppID = appID
	inputMessage.Sender = sender

	message, err := h.app.Services.CreateMessage(inputMessage)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, "message", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// UpdateMessage Updates a message
// @Description Updates a message
// @Tags Admin
// @ID UpdateMessage
// @Accept  json
// @Param data body model.Message true "body json"
// @Success 200 {object} model.Message
// @Security AdminUserAuth
// @Router /admin/message [put]
func (h AdminApisHandler) UpdateMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	/*var message *model.Message
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	if message.ID == nil {
		return l.HTTPResponseErrorData(logutils.StatusMissing, "message id", nil, nil, http.StatusBadRequest, false)
	}

	message.OrgID = claims.OrgID
	message.AppID = claims.AppID

	message, err = h.app.Services.UpdateMessage(&claims.Subject, message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, "message", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data) */

	return l.HTTPResponseError("disabled api", errors.New("disabled api"), 500, true)
}

// GetMessage Retrieves a message by id
// @Description Retrieves a message by id
// @Tags Admin
// @ID GetMessage
// @Param id path string true "id"
// @Accept  json
// @Produce plain
// @Success 200 {object} model.Message
// @Security AdminUserAuth
// @Router /admin/message/{id} [get]
func (h AdminApisHandler) GetMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	message, err := h.app.Services.GetMessage(claims.OrgID, claims.AppID, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "message", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// DeleteMessage Deletes a message with id
// @Description Deletes a message with id
// @Tags Admin
// @ID DeleteMessage
// @Param id path string true "id"
// @Accept  json
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/message/{id} [delete]
func (h AdminApisHandler) DeleteMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Services.DeleteMessage(claims.OrgID, claims.AppID, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "message", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// GetAllAppVersions Gets all available app versions
// @Description Gets all available app versions
// @Tags Admin
// @ID GetAllAppVersions
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/app_versions [get]
func (h AdminApisHandler) GetAllAppVersions(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	appVersions, err := h.app.Services.GetAllAppVersions(claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "app versions", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(appVersions)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// GetAllAppPlatforms Gets all available app platforms
// @Description Gets all available app platforms
// @Tags Admin
// @ID GetAllAppPlatforms
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/app_platforms [get]
func (h AdminApisHandler) GetAllAppPlatforms(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	appPlatforms, err := h.app.Services.GetAllAppPlatforms(claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "app platforms", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(appPlatforms)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// GetMessagesStats gives messages stats
func (h AdminApisHandler) GetMessagesStats(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	//TODO

	list := []Def.AdminResGetMessagesStats{}

	//list = append(list, item1)

	/*	aa := Def.AdminResGetMessagesStats{}

		fieldID := "fdsfds"
		item1 := Def.AdminResGetMessagesStats{FieldId: &fieldID}

		aa = append(aa, struct {
			FieldId *string "json:\"field_id,omitempty\""
		}{}) */

	fieldID := "fdsfds"
	item1 := Def.AdminResGetMessagesStats{FieldId: &fieldID}

	list = append(list, item1)

	data, err := json.Marshal(list)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}
