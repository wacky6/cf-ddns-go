package ddns

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

// CloudFlareProvider encapsulates a CloudFlare DDNS provider.
type CloudFlareProvider struct {
	api *cloudflare.API
}

// CreateCloudFlareProvider creates a DDNS Provider for CloudFlare.
func CreateCloudFlareProvider() (Provider, error) {
	api, err := cloudflare.New(os.Getenv("CF_KEY"), os.Getenv("CF_EMAIL"))

	if err != nil {
		return nil, fmt.Errorf("can't create cloudflare api: %w", err)
	}

	return CloudFlareProvider{api}, nil
}

// VerifyConfig checks the provided credential can access zone information.
func (cf CloudFlareProvider) VerifyConfig() error {
	_, err := cf.api.ListZones()
	if err != nil {
		return fmt.Errorf("can't verify credentials: %w", err)
	}

	return nil
}

// SetRecord sets the `record` for fully qualified domain name `fqdn`
func (cf CloudFlareProvider) SetRecord(fqdn string, record Record) error {
	zones, err := cf.api.ListZones()
	if err != nil {
		return fmt.Errorf("can't list zones: %w", err)
	}

	var zoneID string
	for _, zone := range zones {
		if strings.HasSuffix(fqdn, zone.Name) {
			zoneID = zone.ID
			break
		}
	}

	if len(zoneID) == 0 {
		return fmt.Errorf("fqdn %q is not managed by account", fqdn)
	}

	records, err := cf.api.DNSRecords(zoneID, cloudflare.DNSRecord{Name: fqdn, Type: record.Type})
	if err != nil {
		return fmt.Errorf("can't retrieve dns records: %w", err)
	}

	var recordID string
	var recordContent string
	for _, record := range records {
		if fqdn == record.Name {
			recordID = record.ID
			recordContent = record.Content
		}
	}

	if recordContent == record.Content {
		// Record's content is identical.
		return nil
	}

	if len(recordID) > 0 {
		// Record exists, update it.
		err = cf.api.UpdateDNSRecord(zoneID, recordID, cloudflare.DNSRecord{Content: record.Content})
	} else {
		// Record doesn't exist, create one.
		_, err = cf.api.CreateDNSRecord(zoneID, cloudflare.DNSRecord{Type: record.Type, Content: record.Content, TTL: 120, Proxied: false, Name: fqdn})
	}

	return err
}
