package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
)

type Adapter struct {
	coreURL               string
	serviceAccountManager *authservice.ServiceAccountManager
}

// NewCoreAdapter creates a new adapter for Core API
func NewCoreAdapter(coreURL string, serviceAccountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{coreURL: coreURL, serviceAccountManager: serviceAccountManager}
}

// RetrieveUseAccounts retrieves Core user account based on critera
func (a *Adapter) RetrieveCoreUserAccountByCriteria(accountCriteria map[string]interface{}, appID *string, orgID *string) ([]map[string]interface{}, error) {

	if a.serviceAccountManager == nil {
		log.Println("RetrieveCoreUserAccountByCriteria: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/accounts", accountCriteria)
	queryString := ""
	if appID != nil {
		queryString += "?app_id=" + *appID
	}
	if orgID != nil {
		if queryString == "" {
			queryString += "?"
		} else {
			queryString += "&"
		}
		queryString += "org_id=" + *orgID
	}
	bodyBytes, err := json.Marshal(accountCriteria)
	if err != nil {
		log.Printf("RetrieveCoreUserAccountByCriteria: error marshalling body - %s", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url+queryString, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("RetrieveCoreUserAccountByCriteria: error creating request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	appIDVal := "all"
	if appID != nil {
		appIDVal = *appID
	}
	orgIDVal := "all"
	if orgID != nil {
		appIDVal = *orgID
	}

	resp, err := a.serviceAccountManager.MakeRequest(req, appIDVal, orgIDVal)
	if err != nil {
		log.Printf("RetrieveCoreUserAccountByCriteria: error sending request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("RetrieveCoreUserAccountByCriteria: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("RetrieveCoreUserAccountByCriteria: error with response code != 200")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("RetrieveCoreUserAccountByCriteria: unable to read json: %s", err)
		return nil, fmt.Errorf("RetrieveCoreUserAccountByCriteria: unable to parse json: %s", err)
	}

	var coreAccounts []map[string]interface{}
	err = json.Unmarshal(data, &coreAccounts)
	if err != nil {
		log.Printf("RetrieveCoreUserAccountByCriteria: unable to parse json: %s", err)
		return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
	}

	return coreAccounts, nil

}
