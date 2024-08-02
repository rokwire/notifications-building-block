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
	"bytes"
	"fmt"
	"net/http"
	"notifications/core"
	"notifications/core/model"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/core-auth-library-go/v3/webauth"

	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"

	httpSwagger "github.com/swaggo/http-swagger"
)

// Adapter entity
type Adapter struct {
	host string
	port string

	auth *Auth

	cachedYamlDoc []byte

	apisHandler         ApisHandler
	adminApisHandler    AdminApisHandler
	internalApisHandler InternalApisHandler
	bbsApisHandler      BBsAPIsHandler

	app *core.Application

	corsAllowedOrigins []string
	corsAllowedHeaders []string

	logger *logs.Logger
}

type handlerFunc = func(*logs.Log, *http.Request, *tokenauth.Claims) logs.HTTPResponse

// Start starts the module
func (we Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)

	// handle apis
	baseRouter := router.PathPrefix("/notifications").Subrouter()
	baseRouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	baseRouter.HandleFunc("/doc", we.serveDoc)
	baseRouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version, nil)).Methods("GET")

	mainRouter := baseRouter.PathPrefix("/api").Subrouter()

	// DEPRECATED
	mainRouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	mainRouter.HandleFunc("/doc", we.serveDoc)
	mainRouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version, nil)).Methods("GET")
	//

	// Internal APIs
	// DEPRECATED - Use "bbs" APIs
	mainRouter.HandleFunc("/int/message", we.wrapFunc(we.internalApisHandler.SendMessage, we.auth.internal)).Methods("POST")
	mainRouter.HandleFunc("/int/v2/message", we.wrapFunc(we.internalApisHandler.SendMessageV2, we.auth.internal)).Methods("POST")
	mainRouter.HandleFunc("/int/mail", we.wrapFunc(we.internalApisHandler.SendMail, we.auth.internal)).Methods("POST")

	// Client APIs
	mainRouter.HandleFunc("/token", we.wrapFunc(we.apisHandler.StoreToken, we.auth.client.Standard)).Methods("POST")
	mainRouter.HandleFunc("/user", we.wrapFunc(we.apisHandler.GetUser, we.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/user", we.wrapFunc(we.apisHandler.UpdateUser, we.auth.client.Standard)).Methods("PUT")
	mainRouter.HandleFunc("/user", we.wrapFunc(we.apisHandler.DeleteUser, we.auth.client.Standard)).Methods("DELETE")
	mainRouter.HandleFunc("/messages", we.wrapFunc(we.apisHandler.GetUserMessages, we.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/messages", we.wrapFunc(we.apisHandler.DeleteUserMessages, we.auth.client.Standard)).Methods("DELETE")
	mainRouter.HandleFunc("/messages/read", we.wrapFunc(we.apisHandler.UpdateAllUserMessagesRead, we.auth.client.Standard)).Methods("PUT")
	mainRouter.HandleFunc("/messages/stats", we.wrapFunc(we.apisHandler.GetUserMessagesStats, we.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/message", we.wrapFunc(we.apisHandler.CreateMessage, we.auth.client.Permissions)).Methods("POST")
	mainRouter.HandleFunc("/message/{id}", we.wrapFunc(we.apisHandler.GetUserMessage, we.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/message/{id}", we.wrapFunc(we.apisHandler.DeleteUserMessage, we.auth.client.Standard)).Methods("DELETE")
	mainRouter.HandleFunc("/message/{id}/read", we.wrapFunc(we.apisHandler.UpdateReadMessage, we.auth.client.Standard)).Methods("PUT")
	mainRouter.HandleFunc("/topics", we.wrapFunc(we.apisHandler.GetTopics, we.auth.client.Standard)).Methods("GET")
	//not used and disabled because of the refactoring
	//mainRouter.HandleFunc("/topic/{topic}/messages", we.wrapFunc(we.apisHandler.GetTopicMessages, we.auth.client.Standard)).Methods("GET")
	mainRouter.HandleFunc("/topic/{topic}/subscribe", we.wrapFunc(we.apisHandler.Subscribe, we.auth.client.Standard)).Methods("POST")
	mainRouter.HandleFunc("/topic/{topic}/unsubscribe", we.wrapFunc(we.apisHandler.Unsubscribe, we.auth.client.Standard)).Methods("POST")
	mainRouter.HandleFunc("/push-subscription", we.wrapFunc(we.apisHandler.PushSubscription, we.auth.client.Standard)).Methods("POST")

	// Admin APIs
	adminRouter := mainRouter.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/app-versions", we.wrapFunc(we.adminApisHandler.GetAllAppVersions, we.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/app-platforms", we.wrapFunc(we.adminApisHandler.GetAllAppPlatforms, we.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/topics", we.wrapFunc(we.adminApisHandler.GetTopics, we.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/topic", we.wrapFunc(we.adminApisHandler.UpdateTopic, we.auth.admin.Permissions)).Methods("POST")
	//not used and disabled because of the refactoring
	//adminRouter.HandleFunc("/messages", we.wrapFunc(we.adminApisHandler.GetMessages, we.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/message", we.wrapFunc(we.adminApisHandler.CreateMessage, we.auth.admin.Permissions)).Methods("POST")
	adminRouter.HandleFunc("/message", we.wrapFunc(we.adminApisHandler.UpdateMessage, we.auth.admin.Permissions)).Methods("PUT")
	adminRouter.HandleFunc("/message/{id}", we.wrapFunc(we.adminApisHandler.GetMessage, we.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/message/{id}", we.wrapFunc(we.adminApisHandler.DeleteMessage, we.auth.admin.Permissions)).Methods("DELETE")
	adminRouter.HandleFunc("/messages/stats/source/{source}", we.wrapFunc(we.adminApisHandler.GetMessagesStats, we.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/configs/{id}", we.wrapFunc(we.adminApisHandler.GetConfig, we.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/configs", we.wrapFunc(we.adminApisHandler.GetConfigs, we.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/configs", we.wrapFunc(we.adminApisHandler.CreateConfig, we.auth.admin.Permissions)).Methods("POST")
	adminRouter.HandleFunc("/configs/{id}", we.wrapFunc(we.adminApisHandler.UpdateConfig, we.auth.admin.Permissions)).Methods("PUT")
	adminRouter.HandleFunc("/configs/{id}", we.wrapFunc(we.adminApisHandler.DeleteConfig, we.auth.admin.Permissions)).Methods("DELETE")

	// BB APIs
	bbsRouter := mainRouter.PathPrefix("/bbs").Subrouter()
	bbsRouter.HandleFunc("/messages", we.wrapFunc(we.bbsApisHandler.SendMessages, we.auth.bbs.Permissions)).Methods("POST")
	bbsRouter.HandleFunc("/messages", we.wrapFunc(we.bbsApisHandler.DeleteMessages, we.auth.bbs.Permissions)).Methods("DELETE")
	bbsRouter.HandleFunc("/messages/{message-id}/recipients", we.wrapFunc(we.bbsApisHandler.AddRecipients, we.auth.bbs.Permissions)).Methods("POST")
	bbsRouter.HandleFunc("/messages/{message-id}/recipients", we.wrapFunc(we.bbsApisHandler.DeleteRecipients, we.auth.bbs.Permissions)).Methods("DELETE")

	//deprecated
	bbsRouter.HandleFunc("/message", we.wrapFunc(we.bbsApisHandler.SendMessage, we.auth.bbs.Permissions)).Methods("POST")
	bbsRouter.HandleFunc("/message/{id}", we.wrapFunc(we.bbsApisHandler.DeleteMessage, we.auth.bbs.Permissions)).Methods("DELETE")
	//

	bbsRouter.HandleFunc("/mail", we.wrapFunc(we.bbsApisHandler.SendMail, we.auth.bbs.Permissions)).Methods("POST")

	var handler http.Handler = router
	if len(we.corsAllowedOrigins) > 0 {
		handler = webauth.SetupCORS(we.corsAllowedOrigins, we.corsAllowedHeaders, router)
	}
	we.logger.Fatalf("Error serving: %v", http.ListenAndServe(":"+we.port, handler))
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")

	if we.cachedYamlDoc != nil {
		http.ServeContent(w, r, "", time.Now(), bytes.NewReader([]byte(we.cachedYamlDoc)))
	} else {
		http.ServeFile(w, r, "./driver/web/docs/gen/def.yaml")
	}
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/doc", we.host)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func loadDocsYAML(baseServerURL string) ([]byte, error) {
	data, _ := os.ReadFile("./driver/web/docs/gen/def.yaml")
	// yamlMap := make(map[string]interface{})
	yamlMap := yaml.MapSlice{}
	err := yaml.Unmarshal(data, &yamlMap)
	if err != nil {
		return nil, err
	}

	for index, item := range yamlMap {
		if item.Key == "servers" {
			var serverList []interface{}
			if baseServerURL != "" {
				serverList = []interface{}{yaml.MapSlice{yaml.MapItem{Key: "url", Value: baseServerURL}}}
			}

			item.Value = serverList
			yamlMap[index] = item
			break
		}
	}

	yamlDoc, err := yaml.Marshal(&yamlMap)
	if err != nil {
		return nil, err
	}

	return yamlDoc, nil
}

func (we Adapter) wrapFunc(handler handlerFunc, authorization tokenauth.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logObj := we.logger.NewRequestLog(req)

		logObj.RequestReceived()

		var response logs.HTTPResponse
		if authorization != nil {
			responseStatus, claims, err := authorization.Check(req)
			if err != nil {
				logObj.SendHTTPResponse(w, logObj.HTTPResponseErrorAction(logutils.ActionValidate, logutils.TypeRequest, nil, err, responseStatus, true))
				return
			}

			//do not crash the service if the deprecated internal auth type is used
			if claims != nil {
				logObj.SetContext("account_id", claims.Subject)
			}
			response = handler(logObj, req, claims)
		} else {
			response = handler(logObj, req, nil)
		}

		logObj.SendHTTPResponse(w, response)
		logObj.RequestComplete()
	}
}

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(host string, port string, app *core.Application, config *model.Config, serviceRegManager *authservice.ServiceRegManager,
	corsAllowedOrigins []string, corsAllowedHeaders []string, logger *logs.Logger) Adapter {
	yamlDoc, err := loadDocsYAML(host)
	if err != nil {
		logger.Fatalf("error parsing docs yaml - %s", err.Error())
	}

	auth, err := NewAuth(app, config, serviceRegManager)
	if err != nil {
		logger.Fatalf("error creating auth - %s", err.Error())
	}

	apisHandler := NewApisHandler(app)
	adminApisHandler := NewAdminApisHandler(app)
	internalApisHandler := NewInternalApisHandler(app)
	bbsApisHandler := NewBBsAPIsHandler(app)
	return Adapter{host: host, port: port, cachedYamlDoc: yamlDoc, auth: auth, apisHandler: apisHandler,
		adminApisHandler: adminApisHandler, internalApisHandler: internalApisHandler, bbsApisHandler: bbsApisHandler,
		app: app, corsAllowedOrigins: corsAllowedOrigins, corsAllowedHeaders: corsAllowedHeaders, logger: logger}
}

// AppListener implements core.ApplicationListener interface
type AppListener struct {
	adapter *Adapter
}
