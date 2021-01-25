package corenetworks_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"corenetworks"
	"github.com/libdns/libdns"
)

var (
	user string = ""
	password string = ""
	domain string = ""
)

func TestMain(m *testing.M) {
	user = os.Getenv("LIBDNS_TEST_USER")
	password = os.Getenv("LIBDNS_TEST_PASSWORD")
	domain = os.Getenv("LIBDNS_TEST_DOMAIN")
	
	if len(user) == 0 || len(password) == 0 || len(domain) == 0 {
		fmt.Println("Environment variables LIBDNS_TEST_USER, LIBDNS_TEST_PASSWORD and LIBDNS_TEST_DOMAIN must be set!")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestGetRecords(t *testing.T) {
	p := &corenetworks.Provider{User: user, Password: password,}

	records, err := p.GetRecords(context.TODO(), domain)
	if err != nil {
		t.Log(records)
		t.Fatal(err)
	}

	t.Log(records)
}

func TestAppendRecords(t *testing.T) {
	p := &corenetworks.Provider{User: user, Password: password,}

	appendedRecords, err := p.AppendRecords(context.TODO(), domain, []libdns.Record{
		{
			Name: "test",
			TTL: 300,
			Type: "TXT",
			Value: "test",
		},
	})
	if err != nil {
		t.Log(appendedRecords)
		t.Fatal(err)
	}
}

func TestSetRecords(t *testing.T) {
	p := &corenetworks.Provider{User: user, Password: password,}

	setRecords, err := p.SetRecords(context.TODO(), domain, []libdns.Record{
		{
			Name: "test",
			TTL: 300,
			Type: "TXT",
			Value: "test",
		},
	})
	if err != nil {
		t.Log(setRecords)
		t.Fatal(err)
	}
}

func TestDeleteRecords(t *testing.T) {
	p := &corenetworks.Provider{User: user, Password: password,}

	delRecords, err := p.DeleteRecords(context.TODO(), domain, []libdns.Record{
		{
			Name: "test",
			Value: "test",
		},
	})
	if err != nil {
		t.Log(delRecords)
		t.Fatal(err)
	}
}
