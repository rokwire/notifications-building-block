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
func (h InternalApisHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
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

	message, err = h.app.Services.CreateMessage(nil, message)
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
func (h InternalApisHandler) SendMail(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on reading email data - %s\n", err.Error())
		http.Error(w, fmt.Sprintf("Error on reading email data - %s\n", err.Error()), http.StatusBadRequest)
		return
	}

	var mailRequest *sendMailRequestBody
	err = json.Unmarshal(data, &mailRequest)
	if err != nil {
		log.Printf("Error on unmarshal the email request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.app.Services.SendMail(mailRequest.ToMail, mailRequest.Subject, mailRequest.Body)
	if err != nil {
		log.Printf("Error on sending email: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
