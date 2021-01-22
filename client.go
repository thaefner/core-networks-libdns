// Package corenetworks https://beta.api.core-networks.de/doc/
package corenetworks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/libdns/libdns"
)

// login authentication payload for token
type credentials struct {
	User string `json:"login"`
	Password string `json:"password"`
}

// token session token received by login
type token struct {
	Token  string `json:"token"`
	Expires time.Duration `json:"expires"`
}

// record dns record form 
type record struct {
	Name string `json:"name,omitempty"`
	TTL time.Duration `json:"ttl,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data,omitempty"`
}

// unconform record returned by core-networks record list
type unconformRecord struct {
	Name string `json:"name,omitempty"`
	TTL  string `json:"ttl,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data,omitempty"`
}

// Zone list item
type Zone struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ZoneDetails zone informations
type ZoneDetails struct {
	Active bool `json:"active"`
	DNSSEC bool `json:"dnssec"`
	Master string `json:"master"`
	Name string `json:"name"`
	TSIG tsig `json:"tsig"`
	Type string `json:"type"`
}

type tsig struct {
	Algorithem string `json:"algo"`
	Secret string `json:"secret"`
}

var (
	// CurrentToken latest token set by Login
	currentToken token = token{"",0}
	// tokenExparation time until token expires
	tokenExperation time.Time = time.Now()
)

// doRequest performes http api request and returns http body
func doRequest(p *Provider, req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + currentToken.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {return nil, err}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {return nil, errors.New(resp.Status)}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {return nil, err}

	return respBody, nil
}

// Login gets authentication token with experation time in seconds
func login(ctx context.Context, p *Provider) error {
	loginBody, _ := json.Marshal(credentials{p.User, p.Password})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://beta.api.core-networks.de/auth/token", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {return err}
	
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {return err}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {return errors.New(resp.Status + ": " + string(respBody))}

	var loginRespBody token
	if json.Unmarshal(respBody, &loginRespBody) != nil {return err}
	currentToken = token{
		Token: loginRespBody.Token,
		Expires: time.Duration(loginRespBody.Expires) * time.Second,
	}
	tokenExperation = time.Now().Add(currentToken.Expires)

	return nil
}

// checkToken refresh token if expired
func checkToken(ctx context.Context, p *Provider) error {
	if tokenExperation.Sub(time.Now()) <= 0 {
		return login(ctx, p)
	}
	return nil
}

// convertRecord converts json api record to general libdns record
func convertRecord(domain string, r record) libdns.Record {
	return libdns.Record{
		ID: domain,
		Type: r.Type,
		Name: r.Name,
		Value: r.Data,
		TTL: r.TTL,
	}
}

func convertUnconformRecord(domain string, r unconformRecord) libdns.Record {
	parsedTTL, _ := time.ParseDuration(r.TTL + "s")
	return libdns.Record{
		ID: domain,
		Type: r.Type,
		Name: r.Name,
		Value: r.Data,
		TTL: parsedTTL,
	}
}

// gets all records in domain
func getAllRecords(ctx context.Context, p *Provider, domain string) ([]libdns.Record, error) {
	checkToken(ctx, p)

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://beta.api.core-networks.de/dnszones/%s/records/", domain), nil)
	if err != nil {return nil, err}
	
	respBody, err := doRequest(p, req)
	if err != nil {return nil, err}

	var records = []unconformRecord{}
	err = json.Unmarshal(respBody, &records)
	if err != nil {return nil, err}
	result := []libdns.Record{}
	for _, r := range records {
		result = append(result, convertUnconformRecord(domain, r))
	}
	
	return result, nil
}

//  sets dns a record for a domain
func setRecord(ctx context.Context, p *Provider, domain string, r record) (libdns.Record, error) {
	checkToken(ctx, p)

	recordBody, _ := json.Marshal(r)
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://beta.api.core-networks.de/dnszones/%s/records/", domain), bytes.NewBuffer(recordBody))
	if err != nil {return libdns.Record{}, err}

	_, err = doRequest(p, req)
	if err != nil {return libdns.Record{}, err}

	commit(ctx, p, domain)
	if err != nil {return libdns.Record{}, err}

	return convertRecord(domain, r), nil
}

// deletes a record
func delRecord(ctx context.Context, p *Provider, domain string, r record) (libdns.Record, error) {
	checkToken(ctx, p)

	recordBody, _ := json.Marshal(r)
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://beta.api.core-networks.de/dnszones/%s/records/delete", domain), bytes.NewBuffer(recordBody))
	if err != nil {return libdns.Record{}, err}

	_, err = doRequest(p, req)
	if err != nil {return libdns.Record{}, err}

	commit(ctx, p, domain)
	if err != nil {return libdns.Record{}, err}

	return convertRecord(domain, r), nil
}

// commit changes for immediate update
func commit(ctx context.Context, p *Provider, domain string) error {
	checkToken(ctx, p)

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://beta.api.core-networks.de/dnszones/%s/records/commit", domain), nil)
	if err != nil {return err}

	_, err = doRequest(p, req)
	if err != nil {return err}

	return nil
}

// GetZones all zones of account
func GetZones(ctx context.Context, p *Provider) ([]Zone, error) {
	checkToken(ctx, p)

	req, err := http.NewRequestWithContext(ctx, "GET", "https://beta.api.core-networks.de/dnszones/", nil)
	if err != nil {return nil, err}
	
	respBody, err := doRequest(p, req)
	if err != nil {return nil, err}

	var zones = []Zone{}
	err = json.Unmarshal(respBody, &zones)
	if err != nil {return nil, err}
	
	return zones, nil
}

// GetZone details of zone
func GetZone(ctx context.Context, p *Provider, domain string) (ZoneDetails, error) {
	checkToken(ctx, p)

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://beta.api.core-networks.de/dnszones/%s", domain), nil)
	if err != nil {return ZoneDetails{}, err}
	
	respBody, err := doRequest(p, req)
	if err != nil {return ZoneDetails{}, err}

	var zone = ZoneDetails{}
	if json.Unmarshal(respBody, &zone) != nil {return ZoneDetails{}, err}

	return zone, nil
}