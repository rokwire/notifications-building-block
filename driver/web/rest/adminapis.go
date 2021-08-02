package rest

import (
	"encoding/json"
	"fmt"
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
// @ID GetTopics
// @Produce plain
// @Success 200
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
// @Produce plain
// @Success 200
// @Router /message [post]
func (h AdminApisHandler) SendMessage(user *model.User, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on reading message data - %s\n", err.Error())
		http.Error(w, fmt.Sprintf("Error on reading message data - %s\n", err.Error()), http.StatusBadRequest)
		return
	}

	var message model.Message
	err = json.Unmarshal(data, &message)
	if err != nil {
		log.Printf("Error on unmarshal the message request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.app.Services.SendMessage(message)
	if err != nil {
		log.Printf("Error on sending message: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
