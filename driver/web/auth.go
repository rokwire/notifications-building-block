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
	"net/http"
	"notifications/core"
	"notifications/core/model"

	"github.com/rokwire/core-auth-library-go/v2/authorization"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
)

// Auth handler
type Auth struct {
	client   tokenauth.Handlers
	admin    tokenauth.Handlers
	internal InternalAuth

	logger *logs.Logger
}

// NewAuth creates new auth handler
func NewAuth(app *core.Application, config *model.Config, serviceRegManager *authservice.ServiceRegManager, logger *logs.Logger) (*Auth, error) {
	client, err := newClientAuth(serviceRegManager, logger)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "client auth", nil, err)
	}
	clientHandlers := tokenauth.NewHandlers(client)

	admin, err := newAdminAuth(serviceRegManager, logger)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "admin auth", nil, err)
	}
	adminHandlers := tokenauth.NewHandlers(admin)

	internal := newInternalAuth(config.InternalAPIKey)

	auth := Auth{
		client:   clientHandlers,
		admin:    adminHandlers,
		internal: internal,
		logger:   logger,
	}
	return &auth, nil
}

///////

// InternalAuth handling the internal calls fromother BBs
type InternalAuth struct {
	internalAPIKey string
}

func newInternalAuth(internalAPIKey string) InternalAuth {
	return InternalAuth{internalAPIKey: internalAPIKey}
}

// Check verifies the internal API key
func (auth InternalAuth) Check(req *http.Request) (int, *tokenauth.Claims, error) {
	apiKey := req.Header.Get("INTERNAL-API-KEY")

	//check if there is api key in the header
	if len(apiKey) == 0 {
		//no key, so return 400
		return http.StatusBadRequest, nil, errors.New("Bad Request")
	}

	if auth.internalAPIKey != apiKey {
		//not exist, so return 401
		return http.StatusUnauthorized, nil, errors.New("Unauthorized")
	}

	return http.StatusOK, nil, nil
}

// GetTokenAuth returns nil
func (auth InternalAuth) GetTokenAuth() *tokenauth.TokenAuth {
	return nil
}

// ClientAuth entity
type ClientAuth struct {
	tokenAuth *tokenauth.TokenAuth
	logger    *logs.Logger
}

// Check validates the request contains a valid client token
func (auth ClientAuth) Check(req *http.Request) (int, *tokenauth.Claims, error) {
	claims, err := auth.tokenAuth.CheckRequestTokens(req)
	if err != nil {
		return http.StatusUnauthorized, nil, errors.WrapErrorAction(logutils.ActionValidate, logutils.TypeToken, nil, err)
	}

	if claims.Admin {
		return http.StatusUnauthorized, nil, errors.ErrorData(logutils.StatusInvalid, "admin claim", nil)
	}
	if claims.System {
		return http.StatusUnauthorized, nil, errors.ErrorData(logutils.StatusInvalid, "system claim", nil)
	}

	// TODO: Enable scope authorization
	// err = auth.tokenAuth.AuthorizeRequestScope(claims, req)
	// if err != nil {
	// 	return http.StatusForbidden, nil, errors.WrapErrorAction(logutils.ActionValidate, logutils.TypeScope, nil, err)
	// }

	return http.StatusOK, claims, nil
}

// GetTokenAuth returns the TokenAuth from the handler
func (auth ClientAuth) GetTokenAuth() *tokenauth.TokenAuth {
	return auth.tokenAuth
}

func newClientAuth(serviceRegManager *authservice.ServiceRegManager, logger *logs.Logger) (*ClientAuth, error) {
	servicesTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, nil, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "client token auth", nil, err)
	}

	auth := ClientAuth{tokenAuth: servicesTokenAuth, logger: logger}
	return &auth, nil
}

// AdminAuth entity
type AdminAuth struct {
	tokenAuth *tokenauth.TokenAuth
	logger    *logs.Logger
}

// Check validates the request contains a valid client token
func (auth AdminAuth) Check(req *http.Request) (int, *tokenauth.Claims, error) {
	claims, err := auth.tokenAuth.CheckRequestTokens(req)
	if err != nil {
		return http.StatusUnauthorized, nil, errors.WrapErrorAction(logutils.ActionValidate, logutils.TypeToken, nil, err)
	}

	if !claims.Admin {
		return http.StatusUnauthorized, nil, errors.ErrorData(logutils.StatusInvalid, "admin claim", nil)
	}

	return http.StatusOK, claims, nil
}

// GetTokenAuth returns the TokenAuth from the handler
func (auth AdminAuth) GetTokenAuth() *tokenauth.TokenAuth {
	return auth.tokenAuth
}

func newAdminAuth(serviceRegManager *authservice.ServiceRegManager, logger *logs.Logger) (*AdminAuth, error) {
	adminPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/admin_permission_policy.csv")
	adminTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, adminPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "admin token auth", nil, err)
	}

	auth := AdminAuth{tokenAuth: adminTokenAuth, logger: logger}
	return &auth, nil
}
