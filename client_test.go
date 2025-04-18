package infomaniak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// RoundTripFunc to mock transport layer
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip to allow to use RoundTripFunc as transport layer
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newHttpTestClient returns *http.Client with transport replaced to avoid making real calls
func newHttpTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

// newTestClient returns new client that returns the given answer for an http call
func newTestClient(resultData string, cachedZones *[]IkZone) *Client {
	httpClient := newHttpTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{"result":"success", "data":%s}`, resultData))),
			Header:     make(http.Header),
		}
	})
	return &Client{HttpClient: httpClient, managedZones: cachedZones}
}

// aResponseWithId returns the given ID as the id of a record in form of an http response
func aResponseWithId(id int) *http.Response {
	return aResponse(fmt.Sprintf(`"id": %d`, id))
}

// aResponse returns the given string as the data of a record in form of an http response
func aResponse(recString string) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{"result":"success", "data":{%s}}`, recString))),
		Header:     make(http.Header),
	}
}

func Test_GetInfomaniakManagedZone_ReturnsManagedZoneForDomain(t *testing.T) {
	managedZone := "example.com"
	client := newTestClient(fmt.Sprintf(`[ { "fqdn":"%s" } ]`, managedZone), nil)
	zone, err := client.GetInfomaniakManagedZone(context.TODO(), "subdomain."+managedZone)

	if err != nil {
		t.Fatal(err)
	}

	if zone.Fqdn != managedZone {
		t.Fatalf("Expected zone %s, got %s", managedZone, zone.Fqdn)
	}
}

func Test_GetInfomaniakManagedZone_ReturnsErrorIfZoneNotFound(t *testing.T) {
	managedZone := "example.com"
	client := newTestClient(fmt.Sprintf(`[ { "fqdn":"%s" } ]`, managedZone), nil)
	zone, err := client.GetInfomaniakManagedZone(context.TODO(), "subdomain.test.com")

	if err == nil {
		t.Fatalf("Expected error because no zone matched but got %#v", zone)
	}
}

func Test_GetDnsRecordsForZone_OnlyReturnsRecordsForSpecifiedZone(t *testing.T) {
	domainName := "example.com"
	zone := "subzone." + domainName
	recForZone := IkRecord{ID: 1893, Source: "subzone"}

	jsonString1, err := json.Marshal(recForZone)
	if err != nil {
		t.Fatal(err)
	}
	jsonString2, err := json.Marshal(IkRecord{ID: 335, Source: "."})
	if err != nil {
		t.Fatal(err)
	}

	client := newTestClient(fmt.Sprintf(`[ %s, %s ]`, jsonString1, jsonString2), &[]IkZone{{Fqdn: domainName}})

	recsForZone, err := client.GetDnsRecordsForZone(context.TODO(), zone)
	if err != nil {
		t.Fatal(err)
	}

	if len(recsForZone) != 1 {
		t.Fatalf("Expected %d records, got %d", 1, len(recsForZone))
	}

	if recsForZone[0].ID != recForZone.ID {
		t.Fatalf("Expected records with ID %d, got %d", recForZone.ID, recsForZone[0].ID)
	}
}

func Test_GetDnsRecordsForZone_RemovesExtraQuotesInTargetOfTxtRecord(t *testing.T) {
	domainName := "example.com"
	rec := IkRecord{ID: 1893, Source: ".", Type: "TXT", Target: "\"target_value\""}

	jsonString, _ := json.Marshal(rec)
	client := newTestClient(fmt.Sprintf(`[ %s]`, jsonString), &[]IkZone{{Fqdn: domainName}})
	recs, _ := client.GetDnsRecordsForZone(context.TODO(), domainName)
	if recs[0].Target != "target_value" {
		t.Fatalf("Expected %s as Target, got %s", "target_value", recs[0].Target)
	}
}

func Test_GetDnsRecordsForZone_DoesNotRemoveExtraQuotesInTargetOfRecordOtherThanTxt(t *testing.T) {
	domainName := "example.com"
	rec := IkRecord{ID: 1893, Source: ".", Type: "MX", Target: "\"target_value\""}

	jsonString, _ := json.Marshal(rec)
	client := newTestClient(fmt.Sprintf(`[ %s]`, jsonString), &[]IkZone{{Fqdn: domainName}})
	recs, _ := client.GetDnsRecordsForZone(context.TODO(), domainName)
	if recs[0].Target != "\"target_value\"" {
		t.Fatalf("Expected %s as Target, got %s", "\"target_value\"", recs[0].Target)
	}
}

func Test_GetDnsRecordsForZone_AdjustsSourceRelativeToGivenZone(t *testing.T) {
	domain := "example.com"
	zone := "test." + domain
	rec := IkRecord{ID: 1893, Source: "sub.test"}

	jsonString, _ := json.Marshal(rec)
	client := newTestClient(fmt.Sprintf(`[ %s]`, jsonString), &[]IkZone{{Fqdn: domain}})
	recs, _ := client.GetDnsRecordsForZone(context.TODO(), zone)
	if recs[0].Source != "sub" {
		t.Fatalf("Expected Source %s, got %s", "sub", recs[0].Source)
	}
}

func Test_CreateOrUpdateRecord_ReturnsUpdatedRecord(t *testing.T) {
	id := 984
	httpClient := newHttpTestClient(func(req *http.Request) *http.Response {
		if req.Method != http.MethodPut {
			t.Fatalf("Expected http method %s, got %s", http.MethodPut, req.Method)
		}
		return aResponseWithId(985)
	})

	client := Client{managedZones: &[]IkZone{{Fqdn: "example.com"}}, HttpClient: httpClient}
	rec, _ := client.CreateOrUpdateRecord(context.TODO(), "example.com", IkRecord{ID: id})
	if rec.ID != 985 {
		t.Fatal("Did not return updated / created record")
	}
}

func Test_CreateOrUpdateRecord_CreatesNewRecord(t *testing.T) {
	id := 445
	httpClient := newHttpTestClient(func(req *http.Request) *http.Response {
		if req.Method != http.MethodPost {
			t.Fatalf("Expected http method %s, got %s", http.MethodPost, req.Method)
		}
		return aResponseWithId(id)
	})

	client := Client{managedZones: &[]IkZone{{Fqdn: "example.com"}}, HttpClient: httpClient}
	rec, _ := client.CreateOrUpdateRecord(context.TODO(), "example.com", IkRecord{ID: 0})
	if rec.ID != id {
		t.Fatalf("Expected ID to be %d, got %d", id, rec.ID)
	}
}

func Test_CreateOrUpdateRecord_RemovesExtraQuotesInTargetOfTxtRecord(t *testing.T) {
	httpClient := newHttpTestClient(func(req *http.Request) *http.Response {
		return aResponse(`"id": 123, "type": "TXT", "target": "\"target_val\""`)
	})

	client := Client{managedZones: &[]IkZone{{Fqdn: "example.com"}}, HttpClient: httpClient}
	rec, _ := client.CreateOrUpdateRecord(context.TODO(), "example.com", IkRecord{ID: 0})
	if rec.Target != "target_val" {
		t.Fatalf("Expected Target to be %s, got %s", "target_val", rec.Target)
	}
}

func Test_CreateOrUpdateRecord_DoesNotRemoveExtraQuotesInTargetOfRecordOtherThanTxt(t *testing.T) {
	httpClient := newHttpTestClient(func(req *http.Request) *http.Response {
		return aResponse(`"id": 123, "type": "MX", "target": "\"target_val\""`)
	})

	client := Client{managedZones: &[]IkZone{{Fqdn: "example.com"}}, HttpClient: httpClient}
	rec, _ := client.CreateOrUpdateRecord(context.TODO(), "example.com", IkRecord{ID: 0})
	if rec.Target != "\"target_val\"" {
		t.Fatalf("Expected Target to be %s, got %s", "\"target_val\"", rec.Target)
	}
}

func Test_CreateOrUpdateRecord_AdjustsSourceRelativeToGivenZone(t *testing.T) {
	domain := "example.com"
	zone := "test." + domain
	httpClient := newHttpTestClient(func(req *http.Request) *http.Response {
		return aResponse(`"id": 123, "source": "sub.test"`)
	})

	client := Client{managedZones: &[]IkZone{{Fqdn: domain}}, HttpClient: httpClient}
	rec, _ := client.CreateOrUpdateRecord(context.TODO(), zone, IkRecord{ID: 0})
	if rec.Source != "sub" {
		t.Fatalf("Expected Source %s, got %s", "sub", rec.Source)
	}
}
