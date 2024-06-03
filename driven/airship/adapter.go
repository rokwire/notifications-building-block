package airship

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Adapter is the Airship adapter
type Adapter struct {
	host        string
	bearerToken string
}

// NewAirshipAdapter creates a new Airship adapter instance
func NewAirshipAdapter(host string, bearerToken string) *Adapter {
	return &Adapter{host: host, bearerToken: bearerToken}
}

// SendNotificationToToken sends a notification to an Airship token
func (a *Adapter) SendNotificationToToken(orgID string, appID string, deviceToken string, title string, body string, data map[string]string) error {
	url := fmt.Sprintf(a.host)

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	deviceTypes := [2]string{"ios", "android"}

	audience := map[string]interface{}{
		"device_token": deviceToken,
	}

	iosAlert := map[string]interface{}{
		title: title,
		body:  body,
	}

	iosNotification := map[string]interface{}{
		"alert": iosAlert,
	}

	androidNotification := map[string]interface{}{
		"title": title,
		"alert": body,
	}

	notification := map[string]interface{}{
		"ios":     iosNotification,
		"android": androidNotification,
	}

	//TODO check body for additional urls, localization and additional parameters for notification

	bodyData := map[string]interface{}{
		"device_types": deviceTypes,
		"audience":     audience,
		"notification": notification,
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		log.Printf("error marshalling airship notification request - %s", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("error creating airship notification request - %s", err)
		return err
	}

	bearerToken := a.bearerToken
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error loading airship response data - %s", err)
		return err
	}

	defer resp.Body.Close()

	//TODO save response?
	if resp.StatusCode != 200 {
		log.Printf("error with airship response code - %d", resp.StatusCode)
		return fmt.Errorf("error with airship response code != 200")
	}
	return nil
}
