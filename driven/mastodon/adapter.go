package mastodon

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	typeMastodon logutils.MessageDataType = "mastodon"
)

// Adapter implements the Mastodon interface
type Adapter struct {
	mastodonHost         string
	notificationEndpoint string
}

// PushSubscription creates a push subscription to a mastadon server
func (a *Adapter) PushSubscription(userID string, mastadonToken string) (map[string]interface{}, error) {
	url := fmt.Sprintf(a.mastodonHost)

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := privateKey.PublicKey

	keys := map[string]interface{}{
		"p256dh": publicKey,
		"auth":   privateKey,
	}

	subscription := map[string]interface{}{
		"endpoint": a.notificationEndpoint,
		"keys":     keys,
	}

	alerts := map[string]interface{}{
		"mention": true,
		"follow":  true,
	}

	data := map[string]interface{}{
		"alerts": alerts,
		"policy": "all",
	}

	bodyData := map[string]interface{}{
		"subscription": subscription,
		"data":         data,
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		log.Printf("error creating notification request - %s", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("error creating mastadon request - %s", err)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mastadonToken))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error loading mastadon data - %s", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("error with mastadon response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("error with mastadon response code != 200")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading the body data for the loading mastadon data request - %s", err)
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("error converting data for the loading mastadon data request - %s", err)
		return nil, err
	}

	return result, nil
}

// NewMastodonAdapter creates a new mailer adapter instance
func NewMastodonAdapter(mastodonHost string, notificationEndpoint string) *Adapter {

	return &Adapter{mastodonHost: mastodonHost, notificationEndpoint: notificationEndpoint}
}
