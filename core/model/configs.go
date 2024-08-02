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

package model

import (
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	// TypeConfig configs type
	TypeConfig logutils.MessageDataType = "configs"
	// TypeConfigData config data type
	TypeConfigData logutils.MessageDataType = "config data"
	// TypeEnvConfigData env configs type
	TypeEnvConfigData logutils.MessageDataType = "env config data"

	// ConfigTypeEnv is the Config Type for EnvConfigData
	ConfigTypeEnv string = "env"
	// ConfigTypeApplication is the Config ID for ApplicationConfigData
	ConfigTypeApplication string = "application"
)

// Configs contain generic configs
type Configs struct {
	ID          string      `json:"id" bson:"_id"`
	Type        string      `json:"type" bson:"type"`
	AppID       string      `json:"app_id" bson:"app_id"`
	OrgID       string      `json:"org_id" bson:"org_id"`
	System      bool        `json:"system" bson:"system"`
	Data        interface{} `json:"data" bson:"data"`
	DateCreated time.Time   `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time  `json:"date_updated" bson:"date_updated"`
}

// EnvConfigData contains environment configs for this service
type EnvConfigData struct {
	CORSAllowedOrigins []string `json:"cors_allowed_origins" bson:"cors_allowed_origins"`
	CORSAllowedHeaders []string `json:"cors_allowed_headers" bson:"cors_allowed_headers"`
}

// GetConfigData returns a pointer to the given config's Data as the given type T
func GetConfigData[T ConfigData](c Configs) (*T, error) {
	if data, ok := c.Data.(T); ok {
		return &data, nil
	}
	return nil, errors.ErrorData(logutils.StatusInvalid, TypeConfigData, &logutils.FieldArgs{"type": c.Type})
}

// ConfigData represents any set of data that may be stored in a config
type ConfigData interface {
	EnvConfigData | map[string]interface{}
}
