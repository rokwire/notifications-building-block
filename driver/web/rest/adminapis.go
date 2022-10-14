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

package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"notifications/core"
	"notifications/core/model"

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
func (h AdminApisHandler) GetTopics(user *model.CoreToken, w http.ResponseWriter, r *http.Request) {
	topics, err := h.app.Services.GetTopics(user.OrgID, user.AppID)
	if err != nil {
		log.Printf("Error on retrieving all topics: %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(topics)
	if err != nil {
		log.Println("Error on marshal topics")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// UpdateTopic Updated the topic
// @Description Updated the topic.
// @Tags Admin
// @ID UpdateTopic
// @Param data body model.Topic true "body json"
// @Success 200 {object} model.Topic
// @Security AdminUserAuth
// @Router /admin/topic [put]
func (h AdminApisHandler) UpdateTopic(user *model.CoreToken, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on reading message data - %s\n", err.Error())
		http.Error(w, fmt.Sprintf("Error on reading body data - %s\n", err.Error()), http.StatusBadRequest)
		return
	}

	var topic *model.Topic
	err = json.Unmarshal(data, &topic)
	if err != nil {
		log.Printf("Error on unmarshal the body request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topic.OrgID = user.OrgID
	topic.AppID = user.AppID

	_, err = h.app.Services.UpdateTopic(topic)
	if err != nil {
		log.Printf("Error on update topic (%s): %s\n", topic.Name, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err = json.Marshal(topic)
	if err != nil {
		log.Println("Error on marshal topic")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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
func (h AdminApisHandler) GetMessages(user *model.CoreToken, w http.ResponseWriter, r *http.Request) {
	userIDFilter := getStringQueryParam(r, "user")
	topicFilter := getStringQueryParam(r, "topic")
	offsetFilter := getInt64QueryParam(r, "offset")
	limitFilter := getInt64QueryParam(r, "limit")
	orderFilter := getStringQueryParam(r, "order")
	startDateFilter := getInt64QueryParam(r, "start_date")
	endDateFilter := getInt64QueryParam(r, "end_date")

	messages, err := h.app.Services.GetMessages(user.OrgID, user.AppID, userIDFilter, nil, startDateFilter, endDateFilter, topicFilter, offsetFilter, limitFilter, orderFilter)
	if err != nil {
		log.Printf("Error on getting messages: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if messages == nil {
		messages = []model.Message{}
	}

	data, err := json.Marshal(messages)
	if err != nil {
		log.Printf("Error on marshal messages: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// CreateMessage Creates a message
// @Description Creates a message
// @Tags Admin
// @ID CreateMessage
// @Accept  json
// @Param data body model.Message true "body json"
// @Success 200 {object} model.Message
// @Security AdminUserAuth
// @Router /admin/message [post]
func (h AdminApisHandler) CreateMessage(user *model.CoreToken, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on reading message data - %s\n", err.Error())
		http.Error(w, fmt.Sprintf("Error on reading message data - %s\n", err.Error()), http.StatusBadRequest)
		return
	}

	var message *model.Message
	err = json.Unmarshal(data, &message)
	if err != nil {
		log.Printf("Error on unmarshal the message request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err = h.app.Services.CreateMessage(user, message, false)
	if err != nil {
		log.Printf("Error on create message: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(message)
	if err != nil {
		log.Println("Error on marshal message")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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
func (h AdminApisHandler) UpdateMessage(user *model.CoreToken, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on reading message data - %s\n", err.Error())
		http.Error(w, fmt.Sprintf("Error on reading message data - %s\n", err.Error()), http.StatusBadRequest)
		return
	}

	var message *model.Message
	err = json.Unmarshal(data, &message)
	if err != nil {
		log.Printf("Error on unmarshal the message request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if message.ID == nil {
		log.Printf("Error message doesn't contain ID - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err = h.app.Services.UpdateMessage(user, message)
	if err != nil {
		log.Printf("Error on update message: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(message)
	if err != nil {
		log.Println("Error on marshal message")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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
func (h AdminApisHandler) GetMessage(user *model.CoreToken, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("Message id is required")
		http.Error(w, "Message id is required", http.StatusBadRequest)
		return
	}

	message, err := h.app.Services.GetMessage(id)
	if err != nil {
		log.Printf("Error on get message with id (%s): %s\n", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Println("Error on marshal message")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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
func (h AdminApisHandler) DeleteMessage(user *model.CoreToken, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		log.Println("Message id is required")
		http.Error(w, "Message id is required", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeleteMessage(id)
	if err != nil {
		log.Printf("Error on delete message with id (%s): %s\n", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetAllAppVersions Gets all available app versions
// @Description Gets all available app versions
// @Tags Admin
// @ID GetAllAppVersions
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/app_versions [get]
func (h AdminApisHandler) GetAllAppVersions(_ *model.CoreToken, w http.ResponseWriter, _ *http.Request) {
	appVersions, err := h.app.Services.GetAllAppVersions()
	if err != nil {
		log.Printf("Error on get app versions: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(appVersions)
	if err != nil {
		log.Println("Error on marshal appVersions")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetAllAppPlatforms Gets all available app platforms
// @Description Gets all available app platforms
// @Tags Admin
// @ID GetAllAppPlatforms
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/app_platforms [get]
func (h AdminApisHandler) GetAllAppPlatforms(_ *model.CoreToken, w http.ResponseWriter, _ *http.Request) {
	appPlatforms, err := h.app.Services.GetAllAppPlatforms()
	if err != nil {
		log.Printf("Error on get app platforms: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(appPlatforms)
	if err != nil {
		log.Println("Error on marshal app platforms")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
