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

// doRequest performes http api request and returns http body
func doRequest(p *Provider, req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + p.CurrentToken)

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
	p.CurrentToken = loginRespBody.Token
	p.TokenExperation = time.Now().Add(time.Duration(loginRespBody.Expires) * time.Second)

	return nil
}

// checkToken refresh token if expired
func checkToken(ctx context.Context, p *Provider) error {
	if len(p.CurrentToken) == 0 || p.TokenExperation.Sub(time.Now()) <= 0 {
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
