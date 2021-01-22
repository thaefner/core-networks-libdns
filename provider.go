package corenetworks

import (
	"context"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for core-networks
type Provider struct {
	User string `json:"user"`
	Password string `json:"password"`
	// CurrentToken latest token set by login(should not be set manual)
	CurrentToken string
	// TokenExparation time until token expires(should not be set manual)
	TokenExperation time.Time
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	records, err := getAllRecords(ctx, p, unFQDN(zone))
	if err != nil {return nil, err}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var appendedRecords []libdns.Record

	// TODO: check if record exsists

	return appendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var setRecords []libdns.Record
	
	for _, r := range recs {
		setRecord, err := setRecord(ctx, p, unFQDN(zone), record{
			Name: r.Name,
			TTL: r.TTL,
			Type: r.Type,
			Data: r.Value,
		})
		if err != nil {return setRecords, err}

		setRecords = append(setRecords, setRecord)
	}

	return setRecords, nil
}

// DeleteRecords deletes the records from the zone.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var delRecords []libdns.Record

	for _, r := range recs {
		delRecord, err := delRecord(ctx, p, unFQDN(zone), record{
			Name: r.Name,
			TTL: r.TTL,
			Type: r.Type,
			Data: r.Value,
		})
		if err != nil {return delRecords, err}

		delRecords = append(delRecords, delRecord)
	}

	return delRecords, nil
}

// unFQDN trims any trailing "." from fqdn. core-network's API does not use FQDNs.
func unFQDN(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)