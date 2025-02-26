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

package core

import (
	"log"
	"notifications/driven/core"
	"notifications/driven/mailer"

	"github.com/rokwire/logging-library-go/v2/logs"
)

type storageListener struct {
	app *Application
}

// OnFirebaseConfigurationsUpdated notifies that the firebase configurations have been updated
func (sl *storageListener) OnFirebaseConfigurationsUpdated() {
	log.Println("OnFirebaseConfigurationsUpdated")

	// set the updated firebase configuration in the firebase adapter
	firebaseConfs, err := sl.app.storage.LoadFirebaseConfigurations()
	if err != nil {
		log.Printf("Error getting the firebase configurations when updated - %s", err.Error())
	}

	err = sl.app.firebase.UpdateFirebaseConfigurations(firebaseConfs)
	if err != nil {
		log.Printf("Error setting the firebase configurations when updated - %s", err.Error())
	}
}

// Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Services Services // expose to the drivers adapters
	Admin    Admin    // expose to the drivers adapters
	BBs      BBs      // expose to the drivers adapters
	logger   *logs.Logger

	storage  Storage
	firebase Firebase
	mailer   Mailer
	core     Core

	//gueue logic
	queueLogic queueLogic

	//delete data logic
	deleteDataLogic deleteDataLogic
}

// Start starts the core part of the application
func (app *Application) Start() {
	//set storage listener
	storageListener := storageListener{app: app}
	app.storage.RegisterStorageListener(&storageListener)

	app.queueLogic.start()
	app.deleteDataLogic.start()
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, firebase Firebase, mailer *mailer.Adapter, logger *logs.Logger, core *core.Adapter) *Application {

	timerDone := make(chan bool)
	deleteDataLogic := deleteDataLogic{logger: *logger, coreAdapter: core, storage: storage}

	application := Application{version: version, build: build, storage: storage, firebase: firebase,
		mailer: mailer, logger: logger, core: core, deleteDataLogic: deleteDataLogic}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Admin = &adminImpl{app: &application}
	application.BBs = &bbsImpl{app: &application}
	application.queueLogic = queueLogic{app: &application, logger: logger, storage: storage, firebase: firebase, timerDone: timerDone, airship: airship}

	return &application
}
