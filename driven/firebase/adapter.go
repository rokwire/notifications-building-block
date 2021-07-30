package firebase

import (
	"context"
	"firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"log"
)

// FirebaseAdapter entity
type Adapter struct {
	firebase *firebase.App
}

func NewFirebaseAdapter(authFile string, projectID string) *Adapter {
	conf, err := google.JWTConfigFromJSON([]byte(authFile),
		"https://www.googleapis.com/auth/firebase",
		"https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}

	tokenSource := conf.TokenSource(context.Background())
	creds := google.Credentials{ProjectID: projectID, TokenSource: tokenSource}
	opt := option.WithCredentials(&creds)
	firebaseApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}

	return &Adapter{firebase: firebaseApp}
}

func (fa *Adapter) Start() error {
	return nil
}

func (fa *Adapter) SendNotificationToToken(token string, title string, body string) error {
	ctx := context.Background()
	client, err := fa.firebase.Messaging(ctx)
	if err == nil {
		message := &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
		}
		_, err = client.Send(ctx, message)
		if err != nil {
			err = fmt.Errorf("error while sending notification to token (%s): %w", token, err.Error())
		}
	}
	return err
}

func (fa *Adapter) SendNotificationToTopic(topic string, title string, body string) error {
	ctx := context.Background()
	client, err := fa.firebase.Messaging(ctx)
	if err == nil {
		message := &messaging.Message{
			Topic: topic,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
		}
		_, err = client.Send(ctx, message)
		if err != nil {
			err = fmt.Errorf("error while sending notification to topic (%s): %s", topic, err.Error())
		}
	}
	return err
}

func (fa *Adapter) SubscribeToTopic(token string, topic string) error {
	ctx := context.Background()
	client, err := fa.firebase.Messaging(ctx)
	if err == nil {
		_, err = client.SubscribeToTopic(ctx, []string{token}, topic)
		if err != nil {
			err = fmt.Errorf("error while subscribing to topic (%s): %w", topic, err.Error())
		}
	}
	return err
}

func (fa *Adapter) UnsubscribeToTopic(token string, topic string) error {
	ctx := context.Background()
	client, err := fa.firebase.Messaging(ctx)
	if err == nil {
		_, err = client.UnsubscribeFromTopic(ctx, []string{token}, topic)
		if err != nil {
			err = fmt.Errorf("error while unsubscribing from topic (%s): %w", topic, err.Error())
		}
	}
	return err
}
