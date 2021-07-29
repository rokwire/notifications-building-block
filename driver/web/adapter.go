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
	"log"
	"net/http"
	"notifications/core"
	"notifications/core/model"
	"notifications/driver/web/rest"
	"notifications/utils"

	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

//Adapter entity
type Adapter struct {
	host          string
	port          string
	auth          *Auth
	authorization *casbin.Enforcer

	apisHandler      rest.ApisHandler
	adminApisHandler rest.AdminApisHandler

	app *core.Application
}

// @title Rokwire Content Building Block API
// @description Rokwire Content Building Block API Documentation.
// @version 0.1.0
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /notifications
// @schemes https

// @securityDefinitions.apikey RokwireAuth
// @in header
// @name ROKWIRE-API-KEY

// @securityDefinitions.apikey AdminUserAuth
// @in header (add Bearer prefix to the Authorization value)
// @name Authorization

// @securityDefinitions.apikey AdminGroupAuth
// @in header
// @name GROUP

//Start starts the module
func (we Adapter) Start() {

	we.auth.Start()

	router := mux.NewRouter().StrictSlash(true)

	// handle apis
	mainRouter := router.PathPrefix("/notifications/api").Subrouter()
	mainRouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	mainRouter.HandleFunc("/doc", we.serveDoc)
	mainRouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")

	mainRouter.HandleFunc("/token", we.apiKeyOrTokenWrapFunc(we.apisHandler.StoreFirebaseToken)).Methods("POST")

	mainRouter.HandleFunc("/subscribe", we.adminAppIDTokenAuthWrapFunc(we.apisHandler.Subscribe)).Methods("POST")
	mainRouter.HandleFunc("/unsubscribe", we.adminAppIDTokenAuthWrapFunc(we.apisHandler.Unsubscribe)).Methods("POST")
	mainRouter.HandleFunc("/message", we.apiKeyOrTokenWrapFunc(we.apisHandler.SendMessage)).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+we.port, router))
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./docs/swagger.yaml")
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/notifications/doc", we.host)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func (we Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

type apiKeysAuthFunc = func(*model.User, http.ResponseWriter, *http.Request)

func (we Adapter) apiKeyOrTokenWrapFunc(handler apiKeysAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		apiKeyAuthenticated := we.auth.apiKeyCheck(w, req)
		userAuthenticated, user, _ := we.auth.userCheck(w, req)

		if apiKeyAuthenticated || userAuthenticated {
			handler(user, w, req)
		}
	}
}

type userAuthFunc = func(http.ResponseWriter, *http.Request)

func (we Adapter) userAuthWrapFunc(handler userAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		ok, _, _ := we.auth.userCheck(w, req)
		if !ok {
			return
		}

		handler(w, req)
	}
}

type adminAuthFunc = func(http.ResponseWriter, *http.Request)

func (we Adapter) adminAppIDTokenAuthWrapFunc(handler adminAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		ok, shiboUser := we.auth.adminCheck(w, req)
		if !ok {
			return
		}

		obj := req.URL.Path // the resource that is going to be accessed.
		act := req.Method   // the operation that the user performs on the resource.

		var HasAccess bool = false
		for _, s := range *shiboUser.Membership {
			HasAccess = we.authorization.Enforce(s, obj, act)
			if HasAccess {
				break
			}
		}

		if !HasAccess {
			log.Printf("Access control error - UIN: %s is trying to apply %s operation for %s\n", shiboUser.Uin, act, obj)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		handler(w, req)
	}
}

func (auth *Auth) adminCheck(w http.ResponseWriter, r *http.Request) (bool, *model.User) {
	return auth.adminAuth.check(w, r)
}

func (auth *AdminAuth) check(w http.ResponseWriter, r *http.Request) (bool, *model.User) {
	//1. Get the token from the request
	rawIDToken, tokenType, err := auth.getIDToken(r)
	if err != nil {
		auth.responseBadRequest(w)
		return false, nil
	}

	//3. Validate the token
	idToken, err := auth.verify(*rawIDToken, *tokenType)
	if err != nil {
		log.Printf("error validating token - %s\n", err)

		auth.responseUnauthorized(*rawIDToken, w)
		return false, nil
	}

	//4. Get the user data from the token
	var userData userData
	if err := idToken.Claims(&userData); err != nil {
		log.Printf("error getting user data from token - %s\n", err)

		auth.responseUnauthorized(*rawIDToken, w)
		return false, nil
	}
	//we must have UIuceduUIN
	if userData.UIuceduUIN == nil {
		log.Printf("error - missing uiuceuin data in the token - %s\n", err)

		auth.responseUnauthorized(*rawIDToken, w)
		return false, nil
	}

	shibboAuth := &model.User{Uin: userData.UIuceduUIN, Email: userData.Email,
		Membership: userData.UIuceduIsMemberOf}

	return true, shibboAuth
}

//NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(host string, port string, app *core.Application, appKeys []string, oidcProvider string,
	oidcAppClientID string, adminAppClientID string, adminWebAppClientID string, phoneAuthSecret string,
	authKeys string, authIssuer string, firebaseAuth string, firebaseProjectID string) Adapter {
	auth := NewAuth(app, appKeys, oidcProvider, oidcAppClientID, adminAppClientID, adminWebAppClientID,
		phoneAuthSecret, authKeys, authIssuer)
	authorization := casbin.NewEnforcer("driver/web/authorization_model.conf", "driver/web/authorization_policy.csv")

	apisHandler := rest.NewApisHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app)
	return Adapter{host: host, port: port, auth: auth, authorization: authorization, apisHandler: apisHandler, adminApisHandler: adminApisHandler, app: app}
}

//AppListener implements core.ApplicationListener interface
type AppListener struct {
	adapter *Adapter
}
