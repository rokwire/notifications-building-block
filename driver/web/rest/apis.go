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

package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"notifications/core"
	"notifications/core/model"
)

// ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

// NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) ApisHandler {
	return ApisHandler{app: app}
}

// Version gives the service version
// @Description Gives the service version.
// @Tags Client
// @ID Version
// @Produce plain
// @Success 200
// @Router /version [get]
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

// StoreFirebaseToken Sends a message to a user, list of users or a topic
// @Description Stores a firebase token and maps it to a idToken if presents
// @Tags Client
// @ID SendMessage
// @Produce plain
// @Success 200
// @Router /message [post]
func (h ApisHandler) StoreFirebaseToken(user *model.User, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal token data - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var tokenData bson.M
	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		log.Printf("Error on unmarshal the create student guide request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var token string
	token = tokenData["token"].(string)

	err = h.app.Services.StoreFirebaseToken(token, user)
	if err != nil {
		log.Printf("Error on creating student guide: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Subscribe Subscribes the current user to a topic
// @Description Subscribes the current user to a topic
// @Tags Client
// @ID Subscribe
// @Produce plain
// @Success 200
// @Router /subscribe [post]
func (h ApisHandler) Subscribe(user *model.User, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on reading message data - %s\n", err.Error())
		http.Error(w, fmt.Sprintf("Error on reading message data - %s\n", err.Error()), http.StatusBadRequest)
		return
	}

	var body subscribeTopicBody
	err = json.Unmarshal(data, &body)
	if err != nil {
		log.Printf("Error on unmarshal the message request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(*body.Token) == 0 || len(*body.Topic) == 0 {
		log.Printf("Missing token or topic within the json body")
		http.Error(w, "Missing token or topic within the json body", http.StatusBadRequest)
	}

	err = h.app.Services.SubscribeToTopic(*body.Token, user, *body.Topic)
	if err != nil {
		log.Printf("Error on subscribe to topic (%s): %s\n", *body.Topic, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type subscribeTopicBody struct {
	Token *string `json:"token"`
	Topic *string `json:"topic"`
}

// Unsubscribe Unsubscribes the current user to a topic
// @Description Unsubscribes the current user to a topic
// @Tags Client
// @ID Unsubscribe
// @Produce plain
// @Success 200
// @Router /unsubscribe [post]
func (h ApisHandler) Unsubscribe(user *model.User, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on reading message data - %s\n", err.Error())
		http.Error(w, fmt.Sprintf("Error on reading body data - %s\n", err.Error()), http.StatusBadRequest)
		return
	}

	var body unsubscribeTopicBody
	err = json.Unmarshal(data, &body)
	if err != nil {
		log.Printf("Error on unmarshal the body request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(*body.Token) == 0 || len(*body.Topic) == 0 {
		log.Printf("Missing token or topic within the json body")
		http.Error(w, "Missing token or topic within the json body", http.StatusBadRequest)
	}

	err = h.app.Services.UnsubscribeToTopic(*body.Token, user, *body.Topic)
	if err != nil {
		log.Printf("Error on unsubscribe to topic (%s): %s\n", *body.Topic, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type unsubscribeTopicBody struct {
	Token *string `json:"token"`
	Topic *string `json:"topic"`
}

// GetUserMessages Gets all messages for the user
// @Description Gets all topics
// @Tags Client
// @ID GetUserMessages
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /messages [get]
func (h ApisHandler) GetUserMessages(user *model.User, w http.ResponseWriter, r *http.Request) {
	offsetFilter := getInt64QueryParam(r, "offset")
	limitFilter := getInt64QueryParam(r, "limit")
	orderFilter := getStringQueryParam(r, "order")

	var messages []model.Message
	var err error
	if user != nil {
		messages, err = h.app.Services.GetMessages(user.Uin, user.Email, user.Phone, nil, offsetFilter, limitFilter, orderFilter)
		if err != nil {
			log.Printf("Error on getting user messages: %s", err)
			http.Error(w, fmt.Sprintf("Error on getting user messages: %s", err), http.StatusInternalServerError)
			return
		}
	}
	if messages == nil{
		messages = []model.Message{}
	}

	data, err := json.Marshal(messages)
	if err != nil {
		log.Printf("Error on marshal messages: %s\n", err)
		http.Error(w, fmt.Sprintf("Error on marshal messages: %s\n", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetTopics Gets all topics
// @Description Gets all topics
// @Tags Admin
// @ID GetTopics
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /topics [get]
func (h ApisHandler) GetTopics(user *model.User, w http.ResponseWriter, r *http.Request) {

	topics, err := h.app.Services.GetTopics()
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

// GetTopicMessages Gets all messages for topic
// @Description Gets all messages for topic
// @Tags Client
// @ID GetTopicMessages
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /topic/{topic}/messages [get]
func (h ApisHandler) GetTopicMessages(user *model.User, w http.ResponseWriter, r *http.Request) {
	offsetFilter := getInt64QueryParam(r, "offset")
	limitFilter := getInt64QueryParam(r, "limit")
	orderFilter := getStringQueryParam(r, "order")

	params := mux.Vars(r)
	topic := params["topic"]
	if len(topic) <= 0 {
		log.Println("topic is required")
		http.Error(w, "topic is required", http.StatusBadRequest)
		return
	}

	messages, err := h.app.Services.GetMessages(nil, nil, nil, &topic, offsetFilter, limitFilter, orderFilter)
	if err != nil {
		log.Printf("Error on getting messages: %s", err)
		http.Error(w, fmt.Sprintf("Error on getting messages: %s", err), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(messages)
	if err != nil {
		log.Printf("Error on marshal messages: %s\n", err)
		http.Error(w, fmt.Sprintf("Error on marshal messages: %s\n", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
