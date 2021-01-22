package corenetworks

import (
	"time"
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
	Master string `json:"master,omitempty"`
	Name string `json:"name"`
	TSIG struct {
		Algorithem string `json:"algo"`
		Secret string `json:"secret"`
	} `json:"tsig,omitempty"`
	Type string `json:"type"`
}
