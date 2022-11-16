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
	//	coreAuthPrivateKey := getEnvKey("CORE_AUTH_PRIVATE_KEY", true)
	coreAuthPrivateKey := "-----BEGIN RSA PRIVATE KEY-----\n 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\n -----END RSA PRIVATE KEY-----"
	coreBBHost := getEnvKey("CORE_BB_HOST", true)
	contentServiceURL := getEnvKey("NOTIFICATIONS_SERVICE_URL", true)

	config := &model.Config{
		InternalAPIKey:          internalAPIKey,
		CoreAuthPrivateKey:      coreAuthPrivateKey,
		CoreBBHost:              coreBBHost,
		NotificationsServiceURL: contentServiceURL,
	}

	webAdapter := driver.NewWebAdapter(host, port, application, config)

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
