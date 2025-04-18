package infomaniak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/libdns/libdns"
)

// Client that abstracts and calls infomaniak API
type Client struct {
	// infomaniak API token
	Token string

	// http client used for requests
	HttpClient *http.Client

	// cache of domains registered for the
	// current infomaniak account to prevent
	// that we have to load them for each request
	managedZones *[]IkZone

	// mutex to prevent race conditions
	mu sync.Mutex
}

// GetDnsRecordsForZone loads all dns records for a given zone
func (c *Client) GetDnsRecordsForZone(ctx context.Context, zone string) ([]IkRecord, error) {
	infomaniakManagedZone, err := c.GetInfomaniakManagedZone(ctx, zone)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getRecordsEndpointUrl(infomaniakManagedZone), nil)
	if err != nil {
		return nil, err
	}

	var dnsRecords []IkRecord
	_, err = c.doRequest(req, &dnsRecords)
	if err != nil {
		return nil, err
	}

	zoneRecords := make([]IkRecord, 0)
	for _, rec := range dnsRecords {
		recordFqdn := libdns.AbsoluteName(rec.Source, infomaniakManagedZone.Fqdn)
		if strings.HasSuffix(recordFqdn, zone) {
			rec.Source = libdns.RelativeName(recordFqdn, zone)
			cleanRecordTarget(&rec)
			zoneRecords = append(zoneRecords, rec)
		}
	}
	return zoneRecords, nil
}

// CreateOrUpdateRecord creates a record if its Id property is not set, otherwise it updates the record
func (c *Client) CreateOrUpdateRecord(ctx context.Context, zone string, record IkRecord) (*IkRecord, error) {
	infomaniakManagedZone, err := c.GetInfomaniakManagedZone(ctx, zone)
	if err != nil {
		return nil, err
	}

	recordFqdn := libdns.AbsoluteName(record.Source, zone)
	record.Source = libdns.RelativeName(recordFqdn, infomaniakManagedZone.Fqdn)

	rawJson, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	isNew := record.ID == 0
	var method = http.MethodPost
	var endpoint = getRecordsEndpointUrl(infomaniakManagedZone)

	if !isNew {
		method = http.MethodPut
		endpoint = getRecordEndpointUrl(infomaniakManagedZone, fmt.Sprint(record.ID))
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewBuffer(rawJson))
	if err != nil {
		return nil, err
	}

	var updatedRecord IkRecord
	_, err = c.doRequest(req, &updatedRecord)
	if err != nil {
		return nil, err
	}
	updatedRecord.Source = libdns.RelativeName(libdns.AbsoluteName(updatedRecord.Source, infomaniakManagedZone.Fqdn), zone)
	cleanRecordTarget(&updatedRecord)
	return &updatedRecord, nil
}

// DeleteRecord deletes an existing dns record for a given zone
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordId string) error {
	infomaniakManagedZone, err := c.GetInfomaniakManagedZone(ctx, zone)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, getRecordEndpointUrl(infomaniakManagedZone, recordId), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, nil)
	return err
}

// GetInfomaniakManagedZone looks for the zone that is managed by infomaniak
func (c *Client) GetInfomaniakManagedZone(ctx context.Context, domain string) (IkZone, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.managedZones == nil {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, getZonesEndpointUrl(domain), nil)
		if err != nil {
			return IkZone{}, err
		}
		var zones []IkZone
		_, err = c.doRequest(req, &zones)
		if err != nil {
			return IkZone{}, err
		}
		c.managedZones = &zones
	}
	for _, managedZone := range *c.managedZones {
		if strings.HasSuffix(domain, managedZone.Fqdn) {
			return managedZone, nil
		}
	}
	return IkZone{}, fmt.Errorf("could not find the zone managed by infomaniak for %s", domain)
}

// doRequest performs the API call for the given request req and parses the response's data to the given data struct - if the parameter is not nil
func (c *Client) doRequest(req *http.Request, data any) (*IkResponse, error) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	rawResp, err := c.HttpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer rawResp.Body.Close()

	var resp IkResponse
	err = json.NewDecoder(rawResp.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	if rawResp.StatusCode >= 400 || resp.Result != "success" {
		return nil, fmt.Errorf("got errors: HTTP %d: %+v", rawResp.StatusCode, string(resp.Error))
	}

	if data != nil {
		err = json.Unmarshal(resp.Data, data)
		if err != nil {
			return nil, err
		}
	}

	return &resp, nil
}

const apiBaseUrl = "https://api.infomaniak.com"

// getRecordEndpointUrl returns API endpoint for a specific, already existing record
func getRecordEndpointUrl(zone IkZone, recordId string) string {
	return fmt.Sprintf("%s/%s", getRecordsEndpointUrl(zone), recordId)
}

// getRecordsEndpointUrl returns API endpoint for all records of a zone
func getRecordsEndpointUrl(zone IkZone) string {
	return fmt.Sprintf("%s/2/zones/%s/records", apiBaseUrl, zone.Fqdn)
}

// getRecordsEndpointUrl returns API endpoint for all records of a zone
func getZonesEndpointUrl(domain string) string {
	return fmt.Sprintf("%s/2/domains/%s/zones", apiBaseUrl, domain)
}

// cleanRecordTarget Target of returned record is for some types wrapper in extra quotes
func cleanRecordTarget(record *IkRecord) {
	if record.Type == "TXT" {
		record.Target = record.Target[1 : len(record.Target)-1]
	}
}
