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
	corebb "notifications/driven/core"
	"notifications/driven/firebase"
	"notifications/driven/mailer"
	storage "notifications/driven/storage"
	driver "notifications/driver/web"
	"strconv"
	"strings"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/keys"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/sigauth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/envloader"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
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
	envPrefix := strings.ReplaceAll(strings.ToUpper(serviceID), "-", "_") + "_"

	loggerOpts := logs.LoggerOpts{SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version")}
	loggerOpts.SuppressRequests = append(loggerOpts.SuppressRequests, logs.NewStandardHealthCheckHTTPRequestProperties("notifications/api/version")...)
	logger := logs.NewLogger(serviceID, &loggerOpts)
	envLoader := envloader.NewEnvLoader(Version, logger)

	port := envLoader.GetAndLogEnvVar("PORT", false, false)
	if len(port) == 0 {
		port = "80"
	}

	// mongoDB adapter
	mongoDBAuth := envLoader.GetAndLogEnvVar("MONGO_AUTH", true, true)
	mongoDBName := envLoader.GetAndLogEnvVar("MONGO_DATABASE", true, false)
	mongoTimeout := envLoader.GetAndLogEnvVar("MONGO_TIMEOUT", false, false)
	mtOrgID := envLoader.GetAndLogEnvVar(envPrefix+"MULTI_TENANCY_ORG_ID", true, true)
	mtAppID := envLoader.GetAndLogEnvVar(envPrefix+"MULTI_TENANCY_APP_ID", true, true)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, mtOrgID, mtAppID, logger)
	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	// firebase adapter
	firebaseConfs, err := storageAdapter.LoadFirebaseConfigurations()
	if err != nil {
		log.Fatal("Error loading the firebase confogirations from the storage - " + err.Error())
	}
	firebaseAdapter := firebase.NewFirebaseAdapter()
	err = firebaseAdapter.Start(firebaseConfs)
	if err != nil {
		log.Fatal("Cannot start the Firebase adapter - " + err.Error())
	}

	smtpHost := envLoader.GetAndLogEnvVar("SMTP_HOST", true, false)
	smtpPort := envLoader.GetAndLogEnvVar("SMTP_PORT", true, false)
	smtpUser := envLoader.GetAndLogEnvVar("SMTP_USER", true, true)
	smtpPassword := envLoader.GetAndLogEnvVar("SMTP_PASSWORD", true, true)
	smtpFrom := envLoader.GetAndLogEnvVar("SMTP_EMAIL_FROM", true, true)
	smtpPortNum, _ := strconv.Atoi(smtpPort)
	mailAdapter := mailer.NewMailerAdapter(smtpHost, smtpPortNum, smtpUser, smtpPassword, smtpFrom)

	// web adapter
	host := envLoader.GetAndLogEnvVar("HOST", true, false)
	internalAPIKey := envLoader.GetAndLogEnvVar("INTERNAL_API_KEY", true, true)
	coreBBHost := envLoader.GetAndLogEnvVar("CORE_BB_HOST", true, false)
	notificationsServiceURL := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_URL", true, false)

	authService := auth.Service{
		ServiceID:   serviceID,
		ServiceHost: notificationsServiceURL,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceRegLoader, err := auth.NewRemoteServiceRegLoader(&authService, []string{"auth"})
	if err != nil {
		log.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := auth.NewServiceRegManager(&authService, serviceRegLoader, !strings.HasPrefix(host, "http://localhost"))
	if err != nil {
		log.Fatalf("Error initializing service registration manager: %v", err)
	}

	// Service account
	var serviceAccountManager *auth.ServiceAccountManager

	serviceAccountID := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_ACCOUNT_ID", false, false)
	privKeyRaw := envLoader.GetAndLogEnvVar(envPrefix+"PRIV_KEY", true, true)
	privKeyRaw = strings.ReplaceAll(privKeyRaw, "\\n", "\n")
	privKey, err := keys.NewPrivKey(keys.RS256, privKeyRaw)
	if err != nil {
		logger.Errorf("Error parsing priv key: %v", err)
	} else if serviceAccountID == "" {
		logger.Errorf("Missing service account id")
	} else {
		signatureAuth, err := sigauth.NewSignatureAuth(privKey, serviceRegManager, false, false)
		if err != nil {
			logger.Fatalf("Error initializing signature auth: %v", err)
		}

		serviceAccountLoader, err := auth.NewRemoteServiceAccountLoader(&authService, serviceAccountID, signatureAuth)
		if err != nil {
			logger.Fatalf("Error initializing remote service account loader: %v", err)
		}

		serviceAccountManager, err = auth.NewServiceAccountManager(&authService, serviceAccountLoader)
		if err != nil {
			logger.Fatalf("Error initializing service account manager: %v", err)
		}
	}

	coreAdapter := corebb.NewCoreAdapter(coreBBHost, serviceAccountManager)

	config := &model.Config{
		InternalAPIKey:          internalAPIKey,
		CoreBBHost:              coreBBHost,
		NotificationsServiceURL: notificationsServiceURL,
	}

	// application
	application := core.NewApplication(Version, Build, storageAdapter, firebaseAdapter, mailAdapter, logger, coreAdapter)
	application.Start()

	webAdapter := driver.NewWebAdapter(host, port, application, config, serviceRegManager, logger)

	webAdapter.Start()
}
