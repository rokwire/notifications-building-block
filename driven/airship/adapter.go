package airship

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Adapter struct {
	airshipHost string
}

// NewAirshipAdapter creates a new airship adapter instance
func NewAirshipAdapter(airshipHost string) *Adapter {

	return &Adapter{airshipHost: airshipHost}
}

func (a *Adapter) SendNotificationToToken(orgID string, appID string, deviceToken string, title string, body string, data map[string]string) error {
	url := fmt.Sprintf(a.airshipHost)

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

	andoridNotification := map[string]interface{}{
		"title": title,
		"alert": body,
	}

	notification := map[string]interface{}{
		"ios":     iosNotification,
		"android": andoridNotification,
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

	bearerToken := data["bearer_token"]
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
	} else {
		return nil
	}

}
