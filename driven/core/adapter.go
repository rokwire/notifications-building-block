package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"notifications/core/model"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
)

// Adapter is the adapter for Core BB APIs
type Adapter struct {
	coreURL               string
	serviceAccountManager *authservice.ServiceAccountManager
}

// NewCoreAdapter creates a new adapter for Core API
func NewCoreAdapter(coreURL string, serviceAccountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{coreURL: coreURL, serviceAccountManager: serviceAccountManager}
}

// RetrieveCoreUserAccountByCriteria retrieves Core user account based on criteria
func (a *Adapter) RetrieveCoreUserAccountByCriteria(accountCriteria map[string]interface{}, appID *string, orgID *string) ([]model.CoreAccount, error) {

	if a.serviceAccountManager == nil {
		log.Println("RetrieveCoreUserAccountByCriteria: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	appIDVal := "all"
	if len(*appID) != 0 {
		appIDVal = *appID
	}
	orgIDVal := "all"
	if len(*orgID) != 0 {
		orgIDVal = *orgID
	}

	url := fmt.Sprintf("%s/bbs/accounts", a.coreURL)
	queryString := ""
	if appID != nil {
		queryString += "?app_id=" + appIDVal
	}
	if orgID != nil {
		if queryString == "" {
			queryString += "?"
		} else {
			queryString += "&"
		}
		queryString += "org_id=" + orgIDVal
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

	var coreAccounts []model.CoreAccount
	err = json.Unmarshal(data, &coreAccounts)
	if err != nil {
		log.Printf("RetrieveCoreUserAccountByCriteria: unable to parse json: %s", err)
		return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
	}

	return coreAccounts, nil

}

// LoadDeletedMemberships loads deleted memberships
func (a *Adapter) LoadDeletedMemberships() ([]model.DeletedUserData, error) {

	if a.serviceAccountManager == nil {
		log.Println("LoadDeletedMemberships: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/deleted-memberships?service_id=%s", a.coreURL, a.serviceAccountManager.AuthService.ServiceID)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("delete membership: error creating request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, "all", "all")
	if err != nil {
		log.Printf("LoadDeletedMemberships: error sending request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("LoadDeletedMemberships: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("LoadDeletedMemberships: error with response code != 200")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to read json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	var deletedMemberships []model.DeletedUserData
	err = json.Unmarshal(data, &deletedMemberships)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to parse json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	return deletedMemberships, nil
}
