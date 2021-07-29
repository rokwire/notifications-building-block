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

//FirebaseAdapter entity
type FirebaseAdapter struct {
	firebase *firebase.App
}

func NewFirebaseAdapter(authFile string, projectID string) *FirebaseAdapter {
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

	return &FirebaseAdapter{firebase: firebaseApp}
}

func (fa *FirebaseAdapter) Start() error {
	return nil
}

func (fa *FirebaseAdapter) SendNotificationToToken(token string, title string, body string) error {
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
			log.Fatal("error while sending notification to topic (%s): %s", token, err.Error())
			fmt.Errorf("error while sending notification to token (%s): %s", token, err.Error())
		}
	}
	return err
}

func (fa *FirebaseAdapter) SendNotificationToTopic(topic string, title string, body string) error {
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
			log.Fatalf("error while sending notification to topic (%s): %s", topic, err.Error())
			fmt.Errorf("error while sending notification to topic (%s): %s", topic, err.Error())
		}
	}
	return err
}
