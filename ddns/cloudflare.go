package ddns

import (
	"context"
	"errors"
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
	const tokenEnvName = "CF_TOKEN"
	token := os.Getenv(tokenEnvName)
	if len(token) == 0 {
		return nil, errors.New("CF_TOKEN env variable must be provided")
	}

	api, err := cloudflare.NewWithAPIToken(token)

	if err != nil {
		return nil, fmt.Errorf("can't create cloudflare api: %w", err)
	}

	return CloudFlareProvider{api}, nil
}

// VerifyConfig checks the provided credential can access zone information.
func (cf CloudFlareProvider) VerifyConfig() error {
	_, err := cf.api.ListZones(context.Background())
	if err != nil {
		return fmt.Errorf("can't verify credentials: %w", err)
	}

	return nil
}

// SetRecord sets the `record` for fully qualified domain name `fqdn`
func (cf CloudFlareProvider) SetRecord(fqdn string, record Record) error {
	ctx := context.Background()
	zones, err := cf.api.ListZones(ctx)
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

	zoneIdentifier := cloudflare.ZoneIdentifier(zoneID)

	records, _, err := cf.api.ListDNSRecords(
		ctx,
		zoneIdentifier, cloudflare.ListDNSRecordsParams{
			Name: fqdn,
			Type: record.Type,
		},
	)
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
		_, err = cf.api.UpdateDNSRecord(
			ctx,
			zoneIdentifier,
			cloudflare.UpdateDNSRecordParams{
				ID:      recordID,
				Content: record.Content,
			},
		)
	} else {
		// Record doesn't exist, create one.
		_, err = cf.api.CreateDNSRecord(
			ctx,
			zoneIdentifier,
			cloudflare.CreateDNSRecordParams{
				Type:    record.Type,
				Content: record.Content,
				TTL:     120,
				Proxied: cloudflare.BoolPtr(false),
				Name:    fqdn,
			})
	}

	return err
}
