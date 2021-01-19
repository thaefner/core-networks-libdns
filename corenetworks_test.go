package corenetworks

import (
	"testing"
)

var validationStr string = "test"
var domain string = "example.com"
var user string = "user"
var password string = "password"

func TestLogin(t *testing.T) {
	var err = Login(LoginData{user, password})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(CurrentToken)
}

func TestSetRecord(t *testing.T) {
	TestLogin(t)
	var err = SetRecord(domain, DNSRecord{"_acme-challenge", 300, "TXT", validationStr})
	if(err != nil) {
		t.Log(err)
		t.FailNow()
	}
}

func TestCommit(t *testing.T) {
	TestLogin(t)
	var err = Commit(domain)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}