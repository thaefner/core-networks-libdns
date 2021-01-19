package corenetworks

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// https://beta.api.core-networks.de/doc/
const (
	tokenApiUrl = "https://beta.api.core-networks.de/auth/token"
	dnsApiUrl   = "https://beta.api.core-networks.de/dnszones/"
)

// LoginData passed to Login
type LoginData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Token session token received by Login
type Token struct {
	Token  string `json:"token"`
	Expires int64  `json:"expires"`
}

// DNSRecord dns record form 
type DNSRecord struct {
	Name string `json:"name"`
	TTL  int64  `json:"ttl"`
	Type string `json:"type"`
	Data string `json:"data"`
}

// CurrentToken latest token set by Login
var CurrentToken Token

// Login gets authentication token with experation time in seconds
func Login(login LoginData) error {
	loginBody, _ := json.Marshal(login)
	resp, err := http.Post(tokenApiUrl, "application/json", bytes.NewBuffer(loginBody))
	if err != nil {return err} 
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {return err}
	if resp.StatusCode >= 300 {return errors.New(resp.Status + ": " + string(respBody))}
	var loginRespBody Token
	if json.Unmarshal(respBody, &loginRespBody) != nil {return err}
	CurrentToken = loginRespBody
	return nil
}

// GetZones TODO

// GetZone TODO

// GetRecords TODO

// SetRecord sets dns a record for domain
func SetRecord(domain string, record DNSRecord) error {
	var body, _ = json.Marshal(record)
	req, err := http.NewRequest("POST", dnsApiUrl + domain + "/records/", bytes.NewBuffer(body))
	if err != nil {return err}
	req.Header.Set("Authorization", "Bearer " + CurrentToken.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {return err}
	if resp.StatusCode >= 300 {return errors.New(resp.Status)}
	return nil
}

// DelRecord TODO

// Commit changes for immediate dns update
func Commit(domain string) error {
	req, err := http.NewRequest("POST", dnsApiUrl + domain + "/records/commit", nil)
	req.Header.Set("Authorization", "Bearer " + CurrentToken.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {return err}
	if resp.StatusCode >= 300 {return errors.New(resp.Status)}
	return nil
}


