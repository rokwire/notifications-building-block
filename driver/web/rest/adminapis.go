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
	"strconv"
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
// @ID GetTopics
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/topics [get]
func (h AdminApisHandler) GetTopics(user *model.User, w http.ResponseWriter, r *http.Request) {

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
// @Description Updated the topic
// @Tags Admin
// @ID UpdateTopic
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/topic [put]
func (h AdminApisHandler) UpdateTopic(user *model.User, w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Error on update topic (%s): %s\n", *topic.Name, err)
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
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /message [post]
func (h AdminApisHandler) SendMessage(user *model.User, w http.ResponseWriter, r *http.Request) {
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

// GetMessages Gets all messages. This api may be invoked with different filters in the query string
// @Description Gets all topics
// @Tags Admin
// @ID GetTopics
// @Param uin query string false "uin - filter by uin"
// @Param email query string false "email - filter by email"
// @Param phone query string false "phone - filter by phone"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Pissible values: asc, desc. Default: desc"
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/topics [get]
func (h AdminApisHandler) GetMessages(user *model.User, w http.ResponseWriter, r *http.Request) {
	uinFilter := getStringQueryParam(r, "uin")
	emailFilter := getStringQueryParam(r, "email")
	phoneFilter := getStringQueryParam(r, "phone")
	topicFilter := getStringQueryParam(r, "topic")
	offsetFilter := getInt64QueryParam(r, "offset")
	limitFilter := getInt64QueryParam(r, "limit")
	orderFilter := getStringQueryParam(r, "order")

	messages, err := h.app.Services.GetMessages(uinFilter, emailFilter, phoneFilter, topicFilter, offsetFilter, limitFilter, orderFilter)
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
// @Param data body Message true "body data"
// @Accept  json
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/message [post]
func (h AdminApisHandler) CreateMessage(user *model.User, w http.ResponseWriter, r *http.Request) {
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
// @Param data body Message true "body data"
// @Accept  json
// @Produce plain
// @Success 200
// @Security AdminUserAuth
// @Router /admin/message [put]
func (h AdminApisHandler) UpdateMessage(user *model.User, w http.ResponseWriter, r *http.Request) {
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
// @Success 200
// @Security AdminUserAuth
// @Router /admin/message/{id} [get]
func (h AdminApisHandler) GetMessage(user *model.User, w http.ResponseWriter, r *http.Request) {
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
func (h AdminApisHandler) DeleteMessage(user *model.User, w http.ResponseWriter, r *http.Request) {
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
