package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"notifications/core"
	"notifications/core/model"
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
func (h AdminApisHandler) GetTopics(user *model.ShibbolethUser, w http.ResponseWriter, r *http.Request) {

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

// UpdateTopic Updated the topic
// @Description Updated the topic.
// @Tags Admin
// @ID UpdateTopic
// @Param data body model.Topic true "body json"
// @Success 200 {object} model.Topic
// @Security AdminUserAuth
// @Router /admin/topic [put]
func (h AdminApisHandler) UpdateTopic(user *model.ShibbolethUser, w http.ResponseWriter, r *http.Request) {
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

// SendMessage Sends a message to a user, list of users or a topic
// @Description Sends a message to a user, list of users or a topic
// @Tags Client
// @ID SendMessage
// @Accept  json
// @Param data body model.Message true "body json"
// @Success 200 {object} model.Message
// @Security AdminUserAuth
// @Router /message [post]
func (h AdminApisHandler) SendMessage(user *model.ShibbolethUser, w http.ResponseWriter, r *http.Request) {
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

	message, err = h.app.Services.SendMessage(user, message)
	if err != nil {
		log.Printf("Error on sending message: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(message)
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
// @Success 200 {array} model.Message
// @Security AdminUserAuth
// @Router /admin/messages [get]
func (h AdminApisHandler) GetMessages(user *model.ShibbolethUser, w http.ResponseWriter, r *http.Request) {
	userIDFilter := getStringQueryParam(r, "user")
	topicFilter := getStringQueryParam(r, "topic")
	offsetFilter := getInt64QueryParam(r, "offset")
	limitFilter := getInt64QueryParam(r, "limit")
	orderFilter := getStringQueryParam(r, "order")

	messages, err := h.app.Services.GetMessages(userIDFilter, topicFilter, offsetFilter, limitFilter, orderFilter)
	if err != nil {
		log.Printf("Error on getting messages: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
func (h AdminApisHandler) CreateMessage(user *model.ShibbolethUser, w http.ResponseWriter, r *http.Request) {
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

	message, err = h.app.Services.CreateMessage(message)
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
func (h AdminApisHandler) UpdateMessage(user *model.ShibbolethUser, w http.ResponseWriter, r *http.Request) {
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

	message, err = h.app.Services.UpdateMessage(message)
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
func (h AdminApisHandler) GetMessage(user *model.ShibbolethUser, w http.ResponseWriter, r *http.Request) {
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
func (h AdminApisHandler) DeleteMessage(user *model.ShibbolethUser, w http.ResponseWriter, r *http.Request) {
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
