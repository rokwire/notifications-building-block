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
	"fmt"
	"log"
	"net/http"
	"notifications/core"
	"notifications/core/model"
	Def "notifications/driver/web/docs/gen"
	"strings"
	"time"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"

	"github.com/gorilla/mux"
)

// ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

// NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) ApisHandler {
	return ApisHandler{app: app}
}

type getMessagesRequestBody struct {
	IDs []string `json:"ids"`
} // @name getMessagesRequestBody

type tokenBody struct {
	Token *string `json:"token"`
} // @name tokenBody

// Version gives the service version
// @Description Gives the service version.
// @Tags Client
// @ID Version
// @Produce plain
// @Success 200
// @Security RokwireAuth
// @Router /version [get]
func (h ApisHandler) Version(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	return l.HTTPResponseSuccessMessage(h.app.Services.GetVersion())
}

// StoreToken Sends a message to a user, list of users or a topic
// @Description Stores a token and maps it to a idToken if presents
// @Tags Client
// @ID Token
// @Param data body model.TokenInfo true "body json"
// @Accept  json
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /token [post]
func (h ApisHandler) StoreToken(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var tokenInfo model.TokenInfo
	err := json.NewDecoder(r.Body).Decode(&tokenInfo)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	if tokenInfo.Token == nil || len(*tokenInfo.Token) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "token", logutils.StringArgs("empty or nil"), nil, http.StatusBadRequest, false)
	}

	if tokenInfo.AppVersion == nil || len(*tokenInfo.AppVersion) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "app version", logutils.StringArgs("empty or nil"), nil, http.StatusBadRequest, false)
	}

	if tokenInfo.AppPlatform == nil || len(*tokenInfo.AppPlatform) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, "app platform", logutils.StringArgs("empty or nil"), nil, http.StatusBadRequest, false)
	}

	err = h.app.Services.StoreToken(claims.OrgID, claims.AppID, &tokenInfo, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSave, "token", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// GetUser Gets user record
// @Description Gets user record
// @Tags Client
// @ID User
// @Success 200 {array} model.User
// @Security RokwireAuth UserAuth
// @Router /user [get]
func (h ApisHandler) GetUser(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	userMapping, err := h.app.Services.FindUserByID(claims.OrgID, claims.AppID, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionFind, "user", nil, err, http.StatusInternalServerError, true)
	}

	if userMapping == nil {
		return l.HTTPResponseErrorData(logutils.StatusMissing, "user", nil, nil, http.StatusNotFound, false)
	}

	data, err := json.Marshal(userMapping)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// updateUserRequest Wrapper for update user request body
type updateUserRequest struct {
	NotificationsDisabled bool `json:"notifications_disabled" bson:"notifications_disabled"`
} // @name updateUserRequest

// UpdateUser Updates user record
// @Description Updates user record
// @Tags Client
// @ID User
// @Param data body updateUserRequest true "body json"
// @Success 200 {array} model.User
// @Security RokwireAuth UserAuth
// @Router /user [post]
func (h ApisHandler) UpdateUser(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var bodyData updateUserRequest
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	userMapping, err := h.app.Services.UpdateUserByID(claims.OrgID, claims.AppID, claims.Subject, bodyData.NotificationsDisabled)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, "user", nil, err, http.StatusInternalServerError, true)
	}

	if userMapping == nil {
		return l.HTTPResponseErrorData(logutils.StatusMissing, "user", nil, nil, http.StatusNotFound, false)
	}

	responseData, err := json.Marshal(userMapping)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(responseData)
}

// DeleteUser Deletes user record and unlink all messages
// @Description Deletes user record and unlink all messages
// @Tags Client
// @ID DeleteUser
// @Param data body updateUserRequest true "body json"
// @Success 200 {array} model.User
// @Security RokwireAuth UserAuth
// @Router /user [delete]
func (h ApisHandler) DeleteUser(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	err := h.app.Services.DeleteUserWithID(claims.OrgID, claims.AppID, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "user", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// Subscribe Subscribes the current user to a topic
// @Description Subscribes the current user to a topic
// @Tags Client
// @ID Subscribe
// @Param topic path string true "topic"
// @Param data body tokenBody true "body json"
// @Accept  json
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /topic/{topic}/subscribe [post]
func (h ApisHandler) Subscribe(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	topic := params["topic"]
	if len(topic) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("topic"), nil, http.StatusBadRequest, false)
	}

	//TODO check and only decode
	var body tokenBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	err = h.app.Services.SubscribeToTopic(claims.OrgID, claims.AppID, *body.Token, claims.Subject, claims.Anonymous, topic)
	if err != nil {
		return l.HTTPResponseErrorAction("subscribing", "topic", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// Unsubscribe Unsubscribes the current user to a topic
// @Description Unsubscribes the current user to a topic
// @Tags Client
// @ID Unsubscribe
// @Param topic path string true "topic"
// @Param data body tokenBody true "body json"
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /topic/{topic}/unsubscribe [post]
func (h ApisHandler) Unsubscribe(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	topic := params["topic"]
	if len(topic) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("topic"), nil, http.StatusBadRequest, false)
	}

	//TODO check for ID and decode
	var body tokenBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	err = h.app.Services.UnsubscribeToTopic(claims.OrgID, claims.AppID, *body.Token, claims.Subject, claims.Anonymous, topic)
	if err != nil {
		return l.HTTPResponseErrorAction("unsubscribing", "topic", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// TODO - for now all fields but almost all of them will be removed!
type getUserMessageResponse struct {
	OrgID                     string                    `json:"org_id"`
	AppID                     string                    `json:"app_id"`
	ID                        string                    `json:"id"`
	Priority                  int                       `json:"priority"`
	Subject                   string                    `json:"subject"`
	Sender                    model.Sender              `json:"sender"`
	Body                      string                    `json:"body"`
	Data                      map[string]string         `json:"data"`
	Recipients                []model.MessageRecipient  `json:"recipients"`
	RecipientsCriteriaList    []model.RecipientCriteria `json:"recipients_criteria_list"`
	RecipientAccountCriteria  map[string]interface{}    `json:"recipient_account_criteria"`
	Topic                     *string                   `json:"topic"`
	CalculatedRecipientsCount *int                      `json:"calculated_recipients_count"`
	DateCreated               *time.Time                `json:"date_created"`
	DateUpdated               *time.Time                `json:"date_updated"`
	Time                      time.Time                 `json:"time"`

	Mute bool `json:"mute"`
	Read bool `json:"read"`
}

// GetUserMessages Gets all messages for the user
func (h ApisHandler) GetUserMessages(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	offsetFilter := getInt64QueryParam(r, "offset")
	limitFilter := getInt64QueryParam(r, "limit")
	orderFilter := getStringQueryParam(r, "order")
	startDateFilter := getInt64QueryParam(r, "start_date")
	endDateFilter := getInt64QueryParam(r, "end_date")
	read := getBoolQueryParam(r, "read")
	mute := getBoolQueryParam(r, "mute")

	var messageIDs []string
	var body getMessagesRequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err == nil {
		messageIDs = body.IDs
	}

	recipientsMessages, err := h.app.Services.GetMessagesRecipientsDeep(claims.OrgID, claims.AppID, &claims.Subject, read, mute, messageIDs, startDateFilter, endDateFilter, nil, offsetFilter, limitFilter, orderFilter)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "messages", nil, err, http.StatusInternalServerError, true)
	}
	result := make([]getUserMessageResponse, len(recipientsMessages))
	for i, item := range recipientsMessages {
		message := item.Message

		respItem := getUserMessageResponse{OrgID: message.OrgID, AppID: message.AppID,
			ID: message.ID, Priority: message.Priority, Subject: message.Subject,
			Sender: message.Sender, Body: message.Body, Data: message.Data, Recipients: message.Recipients,
			RecipientsCriteriaList: message.RecipientsCriteriaList, RecipientAccountCriteria: message.RecipientAccountCriteria,
			Topic: message.Topic, CalculatedRecipientsCount: message.CalculatedRecipientsCount,
			DateCreated: message.DateCreated, DateUpdated: message.DateUpdated,
			Mute: item.Mute, Read: item.Read, Time: message.Time}
		result[i] = respItem
	}
	data, err := json.Marshal(result)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// GetUserMessagesStats Count the messages stats
// @Description Count the messages stats.
// @Tags Client
// @ID GetUserMessagesStats
// @Accept  json
// @Success 200
// @Security UserAuth
// @Router /messages/stats[get]
func (h ApisHandler) GetUserMessagesStats(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var err error
	var unreadMessages *model.MessagesStats
	unreadMessages, err = h.app.Services.GetMessagesStats(claims.OrgID, claims.AppID, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "message stats", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(unreadMessages)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// GetTopics Gets all topics
// @Description Gets all topics
// @Tags Client
// @ID GetTopics
// @Success 200 {array} model.Topic
// @Security RokwireAuth
// @Router /topics [get]
func (h ApisHandler) GetTopics(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	topics, err := h.app.Services.GetTopics(claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "topics", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(topics)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// GetTopicMessages Gets all messages for topic
// @Description Gets all messages for topic
// @Tags Client
// @ID GetTopicMessages
// @Param topic path string true "topic"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Param start_date query string false "start_date - Start date filter in milliseconds as an integer epoch value"
// @Param end_date query string false "end_date - End date filter in milliseconds as an integer epoch value"// @Produce plain
// @Success 200 {array} model.Message
// @Security RokwireAuth UserAuth
// @Router /topic/{topic}/messages [get]
func (h ApisHandler) GetTopicMessages(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	return l.HTTPResponseSuccess()
	/*offsetFilter := getInt64QueryParam(r, "offset")
	limitFilter := getInt64QueryParam(r, "limit")
	orderFilter := getStringQueryParam(r, "order")
	startDateFilter := getInt64QueryParam(r, "start_date")
	endDateFilter := getInt64QueryParam(r, "end_date")

	params := mux.Vars(r)
	topic := params["topic"]
	if len(topic) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("topic"), nil, http.StatusBadRequest, false)
	}

	messages, err := h.app.Services.GetMessages(claims.OrgID, claims.AppID, nil, nil, nil, nil, startDateFilter, endDateFilter, &topic, offsetFilter, limitFilter, orderFilter)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "messages", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(messages)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)*/
}

// GetUserMessage Retrieves a message by id
// @Description Retrieves a message by id
// @Tags Client
// @ID GetUserMessage
// @Param id path string true "id"
// @Accept  json
// @Produce plain
// @Success 200 {object} model.Message
// @Security UserAuth
// @Router /message/{id} [get]
func (h ApisHandler) GetUserMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	message, err := h.app.Services.GetUserMessage(claims.OrgID, claims.AppID, id, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "message", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// DeleteUserMessages Removes the current user from the recipient list of all described messages
// @Description Removes the current user from the recipient list of all described messages
// @Tags Client
// @ID DeleteUserMessages
// @Param data body getMessagesRequestBody false "body json of the all message ids that need to be filtered"
// @Accept  json
// @Success 200
// @Security UserAuth
// @Router /messages [delete]
func (h ApisHandler) DeleteUserMessages(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var messageIDs []string
	var body getMessagesRequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err == nil {
		messageIDs = body.IDs
	}

	errStrings := []string{}
	if len(messageIDs) > 0 {
		for _, id := range messageIDs {
			err := h.app.Services.DeleteUserMessage(claims.OrgID, claims.AppID, claims.Subject, id)
			if err != nil {
				errStrings = append(errStrings, fmt.Sprintf("%s\n", err.Error()))
				log.Printf("Error on delete message with id (%s) for recipient (%s): %s\n", id, claims.Subject, err)
			}
		}
	} else {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeRequestBody, logutils.StringArgs("ids"), nil, http.StatusBadRequest, false)
	}
	if len(errStrings) > 0 {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "message", nil, errors.New(strings.Join(errStrings, "")), http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// CreateMessage Creates a message. Message without subject and body will be interpreted as a data massage and it won't be stored in the database
// @Description Creates a message. Message without subject and body will be interpreted as a data massage and it won't be stored in the database
// @Tags Client
// @ID createMessage
// @Accept  json
// @Param data body model.Message true "body json"
// @Success 200 {object} model.Message
// @Security UserAuth
// @Router /message [post]
func (h ApisHandler) CreateMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var inputData Def.SharedReqCreateMessage
	err := json.NewDecoder(r.Body).Decode(&inputData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	orgID := claims.OrgID
	appID := claims.AppID

	sender := model.Sender{Type: "user", User: &model.CoreAccountRef{UserID: claims.Subject, Name: claims.Name}}

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

// DeleteUserMessage Removes the current user from the recipient list of the message
// @Description Removes the current user from the recipient list of the message
// @Tags Client
// @ID DeleteUserMessage
// @Param id path string true "id"
// @Produce plain
// @Success 200
// @Security UserAuth
// @Router /message/{id} [delete]
func (h ApisHandler) DeleteUserMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Services.DeleteUserMessage(claims.OrgID, claims.AppID, claims.Subject, id)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "message", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// UpdateReadMessage marking an "unread" message as "read"
// @Description marking an "unread" message as "read"
// @Tags Client
// @ID UpdateReadMessage
// @Param id path string true "id"
// @Accept  json
// @Success 200 {object} model.Message
// @Security UserAuth
// @Router message/{id}/read [put]
func (h ApisHandler) UpdateReadMessage(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	message, err := h.app.Services.UpdateReadMessage(claims.OrgID, claims.AppID, id, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, "message read", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// updateAllUserMessagesReadRequest Wrapper for update user read flag
type updateAllUserMessagesReadRequest struct {
	Read bool `json:"read"`
} // @name updateAllUserMessagesReadRequest

// UpdateAllUserMessagesRead marking as "unread" or as "read all user messages
// @Description marking as "unread" or as "read all user messages
// @Tags Client
// @ID UpdateAllUserMessagesRead
// @Param data body updateAllUserMessagesReadRequest true "body json"
// @Accept  json
// @Success 200
// @Security UserAuth
// @Router messages/read [put]
func (h ApisHandler) UpdateAllUserMessagesRead(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var body updateAllUserMessagesReadRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, "all messages read", nil, err, http.StatusInternalServerError, true)
	}

	err = h.app.Services.UpdateAllUserMessagesRead(claims.OrgID, claims.AppID, claims.Subject, body.Read)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, "all messages read", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// PushSubscription Subscribes the current user
// @Description Subscribes the current user
// @Tags Client
// @ID PushSubscription
// @Param data body tokenBody true "body json"
// @Accept  json
// @Success 200
// @Security RokwireAuth UserAuth
// @Router /push-subscription [post]
func (h ApisHandler) PushSubscription(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	// params := mux.Vars(r)
	// topic := params["topic"]
	// if len(topic) == 0 {
	// 	return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("topic"), nil, http.StatusBadRequest, false)
	// }

	// parse the token needed for mastadon

	var body tokenBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	if len(*body.Token) == 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeToken, nil, nil, http.StatusBadRequest, false)
	}

	err = h.app.Services.PushSubscription(claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction("subscribing", "topic", nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}
