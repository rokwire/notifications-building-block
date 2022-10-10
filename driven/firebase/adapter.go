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

package firebase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"notifications/core/model"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Adapter entity
type Adapter struct {
	firebaseConfs []model.FirebaseConf

	//key is org_id-app_id construction
	firebaseClients map[string]firebase.App
}

// NewFirebaseAdapter instance a new Firebase adapter
func NewFirebaseAdapter(firebaseConfs []model.FirebaseConf) *Adapter {
	return &Adapter{firebaseConfs: firebaseConfs}
}

// Start starts the firebase adapter
func (fa *Adapter) Start() error {
	//1. check if there are configs data
	firebaseConfs := fa.firebaseConfs
	if len(firebaseConfs) == 0 {
		return errors.New("there is no firebase configurations")
	}

	//2. create a firebase client for every configuration
	for _, current := range firebaseConfs {
		client, err := fa.createFirebaseClient(current)
		if err != nil {
			return err
		}

		key := fmt.Sprintf("%s-%s", current.OrgID, current.AppID)
		fa.firebaseClients[key] = *client
	}
	return nil
}

func (fa *Adapter) createFirebaseClient(data model.FirebaseConf) (*firebase.App, error) {
	conf, err := google.JWTConfigFromJSON([]byte(data.Auth),
		"https://www.googleapis.com/auth/firebase",
		"https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}

	tokenSource := conf.TokenSource(context.Background())
	creds := google.Credentials{ProjectID: data.ProjectID, TokenSource: tokenSource}
	opt := option.WithCredentials(&creds)
	firebaseApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	return firebaseApp, nil
}

// SendNotificationToToken sends a notification to token
func (fa *Adapter) SendNotificationToToken(token string, title string, body string, data map[string]string) error {
	ctx := context.Background()
	client, err := fa.firebase.Messaging(ctx)
	if err == nil {
		message := &messaging.Message{
			Token: token,
			Data:  data,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
		}
		_, err = client.Send(ctx, message)
		if err != nil {
			log.Printf("error while sending notification to token (%s): %s", token, err)
			err = fmt.Errorf("error while sending notification to token (%s): %s", token, err)
		}
	}
	return err
}

// SendNotificationToTopic sends a notification to a topic
func (fa *Adapter) SendNotificationToTopic(topic string, title string, body string, data map[string]string) error {
	ctx := context.Background()
	client, err := fa.firebase.Messaging(ctx)
	if err == nil {
		message := &messaging.Message{
			Topic: topic,
			Data:  data,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
		}
		_, err = client.Send(ctx, message)
		if err != nil {
			err = fmt.Errorf("error while sending notification to topic (%s): %s", topic, err)
		}
	}
	return err
}

// SubscribeToTopic subscribes to a topic
func (fa *Adapter) SubscribeToTopic(token string, topic string) error {
	ctx := context.Background()
	client, err := fa.firebase.Messaging(ctx)
	if err == nil {
		_, err = client.SubscribeToTopic(ctx, []string{token}, topic)
		if err != nil {
			err = fmt.Errorf("error while subscribing to Firebase topic (%s): %s", topic, err)
		}
	}
	return err
}

// UnsubscribeToTopic unsubscribes from a topic
func (fa *Adapter) UnsubscribeToTopic(token string, topic string) error {
	ctx := context.Background()
	client, err := fa.firebase.Messaging(ctx)
	if err == nil {
		_, err = client.UnsubscribeFromTopic(ctx, []string{token}, topic)
		if err != nil {
			err = fmt.Errorf("error while unsubscribing from topic (%s): %s", topic, err)
		}
	}
	return err
}
