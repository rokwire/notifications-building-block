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

	"github.com/rokmetro/auth-library/authservice"
	"github.com/rokmetro/auth-library/tokenauth"
)

//Auth handler
type Auth struct {
	internalAuth *InternalAuth
	coreAuth     *CoreAuth
}

//NewAuth creates new auth handler
func NewAuth(app *core.Application, internalAPIKey string, coreAuthPrivateKey string) *Auth {
	internalAuth := newInternalAuth(internalAPIKey)
	coreAuth := newCoreAuth(app, coreAuthPrivateKey)

	auth := Auth{
		internalAuth: internalAuth,
		coreAuth:     coreAuth,
	}
	return &auth
}

///////

//InternalAuth handling the internal calls fromother BBs
type InternalAuth struct {
	internalAPIKey string
}

func newInternalAuth(internalAPIKey string) *InternalAuth {
	auth := InternalAuth{internalAPIKey: internalAPIKey}
	return &auth
}

func (auth *InternalAuth) check(w http.ResponseWriter, r *http.Request) bool {
	apiKey := r.Header.Get("INTERNAL-API-KEY")
	//check if there is api key in the header
	if len(apiKey) == 0 {
		//no key, so return 400
		log.Println(fmt.Sprintf("400 - Bad Request"))

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return false
	}

	exist := auth.internalAPIKey == apiKey

	if !exist {
		//not exist, so return 401
		log.Println(fmt.Sprintf("401 - Unauthorized for key %s", apiKey))

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return false
	}
	return true
}

////////////////////////////////////

// CoreAuth implementation
type CoreAuth struct {
	app                *core.Application
	tokenAuth          *tokenauth.TokenAuth
	coreAuthPrivateKey *string
}

func newCoreAuth(app *core.Application, coreAuthPrivateKey string) *CoreAuth {

	serviceLoader := authservice.NewRemoteServiceRegLoader("https://api-dev.rokwire.illinois.edu/core/services/auth/service-regs", []string{"core"})
	authService, err := authservice.NewAuthService("notifications", "https://api-dev.rokwire.illinois.edu/notifications", serviceLoader)
	if err != nil {
		log.Fatalf("Error initializing auth service: %v", err)
	}
	tokenAuth, err := tokenauth.NewTokenAuth(true, authService, nil, nil)
	if err != nil {
		log.Fatalf("Error intitializing token auth: %v", err)
	}

	auth := CoreAuth{app: app, tokenAuth: tokenAuth, coreAuthPrivateKey: &coreAuthPrivateKey}
	return &auth
}

func (ca CoreAuth) coreAuthCheck(w http.ResponseWriter, r *http.Request) (bool, *model.CoreToken) {
	claims, err := ca.tokenAuth.CheckRequestTokens(r)
	if err != nil {
		log.Printf("error validate token: %s", err)
		return false, nil
	}

	if claims != nil {
		if claims.Valid() == nil {
			return true, &model.CoreToken{
				UserID:         &claims.Subject,
				ExternalID:     &claims.UID,
				AppID:          &claims.AppID,
				OrganizationID: &claims.OrgID,
				Scope:          &claims.Scope,
				Permissions:    &claims.Permissions,
				Anonymous:      claims.Anonymous,
			}
		}
	}

	return false, nil
}
