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

package main

import (
	"log"
	"notifications/core"
	"notifications/core/model"
	"notifications/driven/airship"
	corebb "notifications/driven/core"
	"notifications/driven/firebase"
	"notifications/driven/mailer"
	storage "notifications/driven/storage"
	driver "notifications/driver/web"
	"os"
	"strconv"
	"strings"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/keys"
	"github.com/rokwire/core-auth-library-go/v3/sigauth"
	"github.com/rokwire/logging-library-go/v2/logs"
)

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}

	serviceID := "notifications"

	loggerOpts := logs.LoggerOpts{SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version")}
	loggerOpts.SuppressRequests = append(loggerOpts.SuppressRequests, logs.NewStandardHealthCheckHTTPRequestProperties("notifications/api/version")...)
	logger := logs.NewLogger(serviceID, &loggerOpts)

	port := getEnvKey("PORT", false)
	if len(port) == 0 {
		port = "80"
	}

	// mongoDB adapter
	mongoDBAuth := getEnvKey("MONGO_AUTH", true)
	mongoDBName := getEnvKey("MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("MONGO_TIMEOUT", false)
	mtOrgID := getEnvKey("NOTIFICATIONS_MULTI_TENANCY_ORG_ID", true)
	mtAppID := getEnvKey("NOTIFICATIONS_MULTI_TENANCY_APP_ID", true)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, mtOrgID, mtAppID, logger)
	err := storageAdapter.Start()
	if err != nil {
		logger.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	// firebase adapter
	firebaseConfs, err := storageAdapter.LoadFirebaseConfigurations()
	_, err = storageAdapter.LoadFirebaseConfigurations()
	if err != nil {
		logger.Fatal("Error loading the firebase configurations from the storage - " + err.Error())
	}
	firebaseAdapter := firebase.NewFirebaseAdapter()
	err = firebaseAdapter.Start(firebaseConfs)
	if err != nil {
		logger.Warn("Cannot start the Firebase adapter - " + err.Error())
	}

	//airship adapter
	airshipHost := getEnvKey("NOTIFICATIONS_AIRSHIP_HOST", false)
	airshipBearerToken := getEnvKey("NOTIFICATIONS_AIRSHIP_BEARER_TOKEN", false)
	airshipAdapter := airship.NewAirshipAdapter(airshipHost, airshipBearerToken)

	smtpHost := getEnvKey("SMTP_HOST", false)
	smtpPort := getEnvKey("SMTP_PORT", false)
	smtpUser := getEnvKey("SMTP_USER", false)
	smtpPassword := getEnvKey("SMTP_PASSWORD", false)
	smtpFrom := getEnvKey("SMTP_EMAIL_FROM", false)
	smtpPortNum, _ := strconv.Atoi(smtpPort)
	mailAdapter := mailer.NewMailerAdapter(smtpHost, smtpPortNum, smtpUser, smtpPassword, smtpFrom)

	// web adapter
	host := getEnvKey("HOST", true)
	internalAPIKey := getEnvKey("INTERNAL_API_KEY", true)
	coreBBHost := getEnvKey("CORE_BB_HOST", true)
	notificationsServiceURL := getEnvKey("NOTIFICATIONS_SERVICE_URL", true)

	authService := authservice.AuthService{
		ServiceID:   serviceID,
		ServiceHost: notificationsServiceURL,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceRegLoader, err := authservice.NewRemoteServiceRegLoader(&authService, []string{"auth"})
	if err != nil {
		logger.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := authservice.NewServiceRegManager(&authService, serviceRegLoader, !strings.HasPrefix(host, "http://localhost"))
	if err != nil {
		logger.Fatalf("Error initializing service registration manager: %v", err)
	}

	//core adapter
	serviceAccountID := getEnvKey("NOTIFICATIONS_SERVICE_ACCOUNT_ID", false)
	privKeyRaw := getEnvKey("NOTIFICATIONS_PRIV_KEY", false)
	var serviceAccountManager *authservice.ServiceAccountManager
	if privKeyRaw != "" {
		privKeyRaw = strings.ReplaceAll(privKeyRaw, "\\n", "\n")
		privKey, err := keys.NewPrivKey(keys.RS256, privKeyRaw)
		if err != nil {
			log.Fatalf("Failed to parse auth priv key: %v", err)
		}
		signatureAuth, err := sigauth.NewSignatureAuth(privKey, serviceRegManager, false, false)
		if err != nil {
			log.Fatalf("Error initializing signature auth: %v", err)
		}

		serviceAccountLoader, err := authservice.NewRemoteServiceAccountLoader(&authService, serviceAccountID, signatureAuth)
		if err != nil {
			log.Fatalf("Error initializing remote service account loader: %v", err)
		}

		serviceAccountManager, err = authservice.NewServiceAccountManager(&authService, serviceAccountLoader)
		if err != nil {
			log.Fatalf("Error initializing service account manager: %v", err)
		}
	}

	coreAdapter := corebb.NewCoreAdapter(coreBBHost, serviceAccountManager)

	config := &model.Config{
		InternalAPIKey:          internalAPIKey,
		CoreBBHost:              coreBBHost,
		NotificationsServiceURL: notificationsServiceURL,
	}

	// application
	application := core.NewApplication(Version, Build, storageAdapter, firebaseAdapter, mailAdapter, logger, coreAdapter, airshipAdapter)
	application.Start()

	webAdapter := driver.NewWebAdapter(host, port, application, config, serviceRegManager, logger)

	webAdapter.Start()
}

func getEnvKey(key string, required bool) string {
	//get from the environment
	value, exist := os.LookupEnv(key)
	if !exist {
		if required {
			log.Fatal("No provided environment variable for " + key)
		}
	}

	return value
}
