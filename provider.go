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
	var domain = unFQDN(zone)

	// check if record exsists
	recordsExist, err := getAllRecords(ctx, p, domain)
	if err != nil {return appendedRecords, err}
	for _, recordToAdd := range recs {
		for _, recordExist := range recordsExist {
			if !(recordExist.Name == recordToAdd.Name && recordExist.Type == recordToAdd.Type) {
				appendedRecord, err := setRecord(ctx, p, domain, record{
					Name: recordToAdd.Name,
					TTL: recordToAdd.TTL,
					Type: recordToAdd.Type,
					Data: recordToAdd.Value,
				})
				if err != nil {continue}

				appendedRecords = append(appendedRecords, appendedRecord)
			}
		}
	}

	// commits for immediate effect
	err = commit(ctx, p, domain)
	if err != nil {return appendedRecords, err}

	return appendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var setRecords []libdns.Record
	var domain = unFQDN(zone)
	
	// sets records and skips fails
	for _, r := range recs {
		setRecord, err := setRecord(ctx, p, domain, record{
			Name: r.Name,
			TTL: r.TTL,
			Type: r.Type,
			Data: r.Value,
		})
		if err != nil {continue}

		setRecords = append(setRecords, setRecord)
	}

	// commits for immediate effect
	err := commit(ctx, p, domain)
	if err != nil {return setRecords, err}

	return setRecords, nil
}

// DeleteRecords deletes the records from the zone.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var deletedRecords []libdns.Record
	var domain = unFQDN(zone)

	// delete records and skips fails
	for _, r := range recs {
		deletedRecord, err := delRecord(ctx, p, domain, record{
			Name: r.Name,
			TTL: r.TTL,
			Type: r.Type,
			Data: r.Value,
		})
		if err != nil {continue}

		deletedRecords = append(deletedRecords, deletedRecord)
	}

	// commits for immediate effect
	err := commit(ctx, p, domain)
	if err != nil {return deletedRecords, err}

	return deletedRecords, nil
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