package airship

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type m map[string]interface{}

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

	//TODO check body for additional urls, localization and additional parameters for notification

	ios := m{
		"alert": m{
			"title": title,
			"body":  body,
		},
	}
	android := m{
		"title": title,
		"alert": body,
	}

	if val, ok := data["url"]; ok {
		actions := m{
			"open": m{
				"type":    "url",
				"content": val,
			},
		}
		ios["actions"] = actions
		android["actions"] = actions
	}

	bodyData := m{
		"device_types": []string{"ios", "android"},
		"audience": m{
			"channel": deviceToken,
		},
		"notification": m{
			"ios":     ios,
			"android": android,
		},
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
