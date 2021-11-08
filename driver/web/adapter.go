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

package web

import (
	"fmt"
	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"notifications/core"
	"notifications/core/model"
	"notifications/driver/web/rest"
	"notifications/utils"
	"strings"
)

// Adapter entity
type Adapter struct {
	host          string
	port          string
	auth          *Auth
	authorization *casbin.Enforcer

	apisHandler         rest.ApisHandler
	adminApisHandler    rest.AdminApisHandler
	internalApisHandler rest.InternalApisHandler

	app *core.Application
}

// @title Rokwire Notifications Building Block API
// @description Rokwire Notifications Building Block API Documentation.
// @version 0.1.18
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /notifications/api
// @schemes https

// @securityDefinitions.apikey RokwireAuth
// @in header
// @name ROKWIRE-API-KEY

// @securityDefinitions.apikey InternalAuth
// @in header
// @name INTERNAL-API-KEY

// @securityDefinitions.apikey UserAuth
// @in header (add core access token with Bearer prefix to the Authorization value. The token must represent either anonymous or authenticated user )
// @name Authorization

// @securityDefinitions.apikey AdminUserAuth
// @in header (add admin core access token with Bearer prefix to the Authorization value. The token must contain notifications_admin as a permission)
// @name Authorization

// Start starts the module
func (we Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)

	// handle apis
	mainRouter := router.PathPrefix("/notifications/api").Subrouter()
	mainRouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	mainRouter.HandleFunc("/doc", we.serveDoc)
	mainRouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")

	// Internal APIs
	mainRouter.HandleFunc("/int/message", we.internalAPIKeyAuthWrapFunc(we.internalApisHandler.SendMessage)).Methods("POST")

	// Client APIs
	mainRouter.HandleFunc("/token", we.coreWrapFunc(we.apisHandler.StoreFirebaseToken)).Methods("POST")
	mainRouter.HandleFunc("/user", we.coreWrapFunc(we.apisHandler.GetUser)).Methods("GET")
	mainRouter.HandleFunc("/messages", we.coreWrapFunc(we.apisHandler.GetUserMessages)).Methods("GET")
	mainRouter.HandleFunc("/messages", we.coreWrapFunc(we.apisHandler.DeleteUserMessages)).Methods("DELETE")
	mainRouter.HandleFunc("/message", we.coreWrapFunc(we.apisHandler.CreateMessage)).Methods("POST")
	mainRouter.HandleFunc("/message/{id}", we.coreWrapFunc(we.apisHandler.GetMessage)).Methods("GET")
	mainRouter.HandleFunc("/message/{id}", we.coreWrapFunc(we.apisHandler.DeleteUserMessage)).Methods("DELETE")
	mainRouter.HandleFunc("/topics", we.coreWrapFunc(we.apisHandler.GetTopics)).Methods("GET")
	mainRouter.HandleFunc("/topic/{topic}/messages", we.coreWrapFunc(we.apisHandler.GetTopicMessages)).Methods("GET")
	mainRouter.HandleFunc("/topic/{topic}/subscribe", we.coreWrapFunc(we.apisHandler.Subscribe)).Methods("POST")
	mainRouter.HandleFunc("/topic/{topic}/unsubscribe", we.coreWrapFunc(we.apisHandler.Unsubscribe)).Methods("POST")

	// Admin APIs
	adminRouter := mainRouter.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/app-versions", we.coreAdminWrapFunc(we.adminApisHandler.GetAllAppVersions)).Methods("GET")
	adminRouter.HandleFunc("/app-platforms", we.coreAdminWrapFunc(we.adminApisHandler.GetAllAppPlatforms)).Methods("GET")
	adminRouter.HandleFunc("/topics", we.coreAdminWrapFunc(we.adminApisHandler.GetTopics)).Methods("GET")
	adminRouter.HandleFunc("/topic", we.coreAdminWrapFunc(we.adminApisHandler.UpdateTopic)).Methods("POST")
	adminRouter.HandleFunc("/messages", we.coreAdminWrapFunc(we.adminApisHandler.GetMessages)).Methods("GET")
	adminRouter.HandleFunc("/message", we.coreAdminWrapFunc(we.adminApisHandler.CreateMessage)).Methods("POST")
	adminRouter.HandleFunc("/message", we.coreAdminWrapFunc(we.adminApisHandler.UpdateMessage)).Methods("PUT")
	adminRouter.HandleFunc("/message/{id}", we.coreAdminWrapFunc(we.adminApisHandler.GetMessage)).Methods("GET")
	adminRouter.HandleFunc("/message/{id}", we.coreAdminWrapFunc(we.adminApisHandler.DeleteMessage)).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":"+we.port, router))
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./docs/swagger.yaml")
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/notifications/api/doc", we.host)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func (we Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

type coreAuthFunc = func(*model.CoreToken, http.ResponseWriter, *http.Request)

func (we Adapter) coreWrapFunc(handler coreAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		authenticated, user := we.auth.coreAuth.coreAuthCheck(w, req)

		if authenticated {
			handler(user, w, req)
			return
		}
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

type coreAdminAuthFunc = func(*model.CoreToken, http.ResponseWriter, *http.Request)

func (we Adapter) coreAdminWrapFunc(handler coreAdminAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		authenticated, user := we.auth.coreAuth.coreAuthCheck(w, req)

		if authenticated {
			obj := req.URL.Path // the resource that is going to be accessed.
			act := req.Method   // the operation that the user performs on the resource.
			permissions := strings.Split(*user.Permissions, ",")

			HasAccess := false
			for _, s := range permissions {
				HasAccess = we.authorization.Enforce(s, obj, act)
				if HasAccess {
					break
				}
			}
			if HasAccess {
				handler(user, w, req)
				return
			}
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

type internalAPIKeyAuthFunc = func(http.ResponseWriter, *http.Request)

func (we Adapter) internalAPIKeyAuthWrapFunc(handler internalAPIKeyAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		apiKeyAuthenticated := we.auth.internalAuth.check(w, req)

		if apiKeyAuthenticated {
			handler(w, req)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(host string, port string, app *core.Application, config *model.Config) Adapter {
	auth := NewAuth(app, config)
	authorization := casbin.NewEnforcer("driver/web/authorization_model.conf", "driver/web/authorization_policy.csv")

	apisHandler := rest.NewApisHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app)
	internalApisHandler := rest.NewInternalApisHandler(app)
	return Adapter{host: host, port: port, auth: auth, authorization: authorization,
		apisHandler: apisHandler, adminApisHandler: adminApisHandler, internalApisHandler: internalApisHandler, app: app}
}

// AppListener implements core.ApplicationListener interface
type AppListener struct {
	adapter *Adapter
}
