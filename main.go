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
	"notifications/driven/firebase"
	"notifications/driven/mailer"
	storage "notifications/driven/storage"
	driver "notifications/driver/web"
	"os"
	"strconv"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
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

	port := getEnvKey("PORT", true)

	// mongoDB adapter
	mongoDBAuth := getEnvKey("MONGO_AUTH", true)
	mongoDBName := getEnvKey("MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("MONGO_TIMEOUT", false)
	mtOrgID := getEnvKey("NOTIFICATIONS_MULTI_TENANCY_ORG_ID", true)
	mtAppID := getEnvKey("NOTIFICATIONS_MULTI_TENANCY_APP_ID", true)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, mtOrgID, mtAppID)
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

	smtpHost := getEnvKey("SMTP_HOST", true)
	smtpPort := getEnvKey("SMTP_PORT", true)
	smtpUser := getEnvKey("SMTP_USER", true)
	smtpPassword := getEnvKey("SMTP_PASSWORD", true)
	smtpFrom := getEnvKey("SMTP_EMAIL_FROM", true)
	smtpPortNum, _ := strconv.Atoi(smtpPort)
	mailAdapter := mailer.NewMailerAdapter(smtpHost, smtpPortNum, smtpUser, smtpPassword, smtpFrom)

	// application
	application := core.NewApplication(Version, Build, storageAdapter, firebaseAdapter, mailAdapter)
	application.Start()

	// web adapter
	host := getEnvKey("HOST", true)
	internalAPIKey := getEnvKey("INTERNAL_API_KEY", true)
	coreAuthPrivateKey := getEnvKey("CORE_AUTH_PRIVATE_KEY", true)
	coreBBHost := getEnvKey("CORE_BB_HOST", true)
	notificationsServiceURL := getEnvKey("NOTIFICATIONS_SERVICE_URL", true)

	authService := authservice.AuthService{
		ServiceID:   "notifications",
		ServiceHost: notificationsServiceURL,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceRegLoader, err := authservice.NewRemoteServiceRegLoader(&authService, []string{"auth"})
	if err != nil {
		log.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := authservice.NewServiceRegManager(&authService, serviceRegLoader)
	if err != nil {
		log.Fatalf("Error initializing service registration manager: %v", err)
	}

	config := &model.Config{
		InternalAPIKey:          internalAPIKey,
		CoreAuthPrivateKey:      coreAuthPrivateKey,
		CoreBBHost:              coreBBHost,
		NotificationsServiceURL: notificationsServiceURL,
	}

	webAdapter := driver.NewWebAdapter(host, port, application, config, serviceRegManager)

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
