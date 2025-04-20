package infomaniak

import (
	"bytes"
	"context"
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

// aTestHttpClient returns *http.Client with transport replaced to avoid making real calls
func aTestHttpClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

// aTestClient returns new client that returns the given answer for an http call
func aTestClient(resultData string) *Client {
	httpClient := aTestHttpClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{"result":"success", "data":%s}`, resultData))),
			Header:     make(http.Header),
		}
	})
	return &Client{HttpClient: httpClient}
}

// aFailingTestClient returns new client that returns some kind of error when the API is called
func aFailingTestClient(statusCode int, err string) *Client {
	httpClient := aTestHttpClient(func(req *http.Request) *http.Response {
		body := ""
		if statusCode == 200 {
			body = fmt.Sprintf(`{"result":"error", "error":%s}`, err)
		}

		return &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}
	})
	return &Client{HttpClient: httpClient}
}

// aRequestCapturingTestClient returns new client that allows to capture the request parameters
func aRequestCapturingTestClient(resultData string, request *http.Request) *Client {
	httpClient := aTestHttpClient(func(req *http.Request) *http.Response {
		*request = *req
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{"result":"success", "data":%s}`, resultData))),
			Header:     make(http.Header),
		}
	})
	return &Client{HttpClient: httpClient}
}

func Test_GetFqdnOfZoneForDomain_CallsInfomaniakEndpointWithAuthHeader(t *testing.T) {
	zone := "sub.example.com"
	var request http.Request
	client := aRequestCapturingTestClient(fmt.Sprintf(`[ { "fqdn":"%s" } ]`, zone), &request)
	client.Token = "test-token"

	client.GetFqdnOfZoneForDomain(context.TODO(), zone)

	authHeader := request.Header.Get("Authorization")
	if authHeader != "Bearer test-token" {
		t.Fatalf("Authorization header not correct, expected: \"%s\", actual: \"%s\"", "Bearer test-token", authHeader)
	}
}

func Test_GetFqdnOfZoneForDomain_CallsInfomaniakEndpointWithContentTypeHeader(t *testing.T) {
	zone := "sub.example.com"
	var request http.Request
	client := aRequestCapturingTestClient(fmt.Sprintf(`[ { "fqdn":"%s" } ]`, zone), &request)

	client.GetFqdnOfZoneForDomain(context.TODO(), zone)

	contentTypeHeader := request.Header.Get("Content-Type")
	if contentTypeHeader != "application/json" {
		t.Fatalf("Content-Type header not correct, expected: \"%s\", actual: \"%s\"", "application/json", contentTypeHeader)
	}
}

func Test_GetFqdnOfZoneForDomain_CallsInfomaniakEndpointWithGetMethod(t *testing.T) {
	zone := "sub.example.com"
	var request http.Request
	client := aRequestCapturingTestClient(fmt.Sprintf(`[ { "fqdn":"%s" } ]`, zone), &request)

	client.GetFqdnOfZoneForDomain(context.TODO(), zone)

	if request.Method != http.MethodGet {
		t.Fatalf("Wrong http method used, expected: \"%s\", actual: \"%s\"", http.MethodGet, request.Method)
	}
}

func Test_GetFqdnOfZoneForDomain_CallsCorrectInfomaniakEndpoint(t *testing.T) {
	zone := "sub.example.com"
	expectedEndpoint := fmt.Sprintf("https://api.infomaniak.com/2/domains/%s/zones", zone)
	var request http.Request
	client := aRequestCapturingTestClient(fmt.Sprintf(`[ { "fqdn":"%s" } ]`, zone), &request)

	client.GetFqdnOfZoneForDomain(context.TODO(), zone)

	endpoint := request.URL.String()
	if endpoint != expectedEndpoint {
		t.Fatalf("Wrong endpoint used, expected: \"%s\", actual: \"%s\"", expectedEndpoint, endpoint)
	}
}

func Test_GetFqdnOfZoneForDomain_ReturnsManagedZoneForDomain(t *testing.T) {
	managedZone := "example.com"
	client := aTestClient(fmt.Sprintf(`[ { "fqdn":"%s" } ]`, managedZone))

	zone, _ := client.GetFqdnOfZoneForDomain(context.TODO(), "subdomain."+managedZone)

	if zone != managedZone {
		t.Fatalf("Expected zone %s, got %s", managedZone, zone)
	}
}

func Test_GetFqdnOfZoneForDomain_ReturnsMostAccurateManagedZoneForDomain(t *testing.T) {
	managedTopLevelZone := "example.com"
	managedSubZone := "sub.example.com"
	client := aTestClient(fmt.Sprintf(`[ { "fqdn":"%s" }, { "fqdn":"%s" } ]`, managedTopLevelZone, managedSubZone))

	zone, _ := client.GetFqdnOfZoneForDomain(context.TODO(), "test."+managedSubZone)

	if zone != managedSubZone {
		t.Fatalf("Expected zone %s, got %s", managedSubZone, zone)
	}
}

func Test_GetFqdnOfZoneForDomain_ReturnsErrorIfZoneNotFound(t *testing.T) {
	managedZone := "example.com"
	client := aTestClient(fmt.Sprintf(`[ { "fqdn":"%s" } ]`, managedZone))

	zone, err := client.GetFqdnOfZoneForDomain(context.TODO(), "subdomain.test.com")

	if err == nil {
		t.Fatalf("Expected error because no zone matched but got %s", zone)
	}
}

func Test_GetFqdnOfZoneForDomain_ReturnsErrorIfApiCallFails(t *testing.T) {
	client := aFailingTestClient(500, "")

	_, err := client.GetFqdnOfZoneForDomain(context.TODO(), "subdomain.test.com")

	if err == nil {
		t.Fatalf("Expected error because API call failed")
	}
}

func Test_GetFqdnOfZoneForDomain_ReturnsErrorIfApiReturnsError(t *testing.T) {
	client := aFailingTestClient(200, "some error message")

	_, err := client.GetFqdnOfZoneForDomain(context.TODO(), "subdomain.test.com")

	if err == nil {
		t.Fatalf("Expected error because infomaniak API call returned error")
	}
}

func Test_GetDnsRecordsForZone_CallsInfomaniakEndpointWithAuthHeader(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient("[]", &request)
	client.Token = "test-token"

	client.GetDnsRecordsForZone(context.TODO(), "sub.example.com")

	authHeader := request.Header.Get("Authorization")
	if authHeader != "Bearer test-token" {
		t.Fatalf("Authorization header not correct, expected: \"%s\", actual: \"%s\"", "Bearer test-token", authHeader)
	}
}

func Test_GetDnsRecordsForZone_CallsInfomaniakEndpointWithContentTypeHeader(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient("[]", &request)

	client.GetDnsRecordsForZone(context.TODO(), "sub.example.com")

	contentTypeHeader := request.Header.Get("Content-Type")
	if contentTypeHeader != "application/json" {
		t.Fatalf("Content-Type header not correct, expected: \"%s\", actual: \"%s\"", "application/json", contentTypeHeader)
	}
}

func Test_GetDnsRecordsForZone_CallsInfomaniakEndpointWithGetMethod(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient("[]", &request)

	client.GetDnsRecordsForZone(context.TODO(), "example.com")

	if request.Method != http.MethodGet {
		t.Fatalf("Wrong http method used, expected: \"%s\", actual: \"%s\"", http.MethodGet, request.Method)
	}
}

func Test_GetDnsRecordsForZone_CallsCorrectInfomaniakEndpoint(t *testing.T) {
	zone := "sub.example.com"
	expectedEndpoint := fmt.Sprintf("https://api.infomaniak.com/2/zones/%s/records?with=records_description", zone)
	var request http.Request
	client := aRequestCapturingTestClient("[]", &request)

	client.GetDnsRecordsForZone(context.TODO(), zone)

	endpoint := request.URL.String()
	if endpoint != expectedEndpoint {
		t.Fatalf("Wrong endpoint used, expected: \"%s\", actual: \"%s\"", expectedEndpoint, endpoint)
	}
}

func Test_GetDnsRecordsForZone_ParsesNSRecordCorrectly(t *testing.T) {
	client := aTestClient(`[{"id":25,"source":".","type":"NS","ttl":3600,"target":"ns11.infomaniak.ch","updated_at":1659958248}]`)

	res, _ := client.GetDnsRecordsForZone(context.TODO(), "example.com")

	assertEqualsInt(t, "ID", 25, res[0].ID)
	assertEquals(t, "Source", ".", res[0].Source)
	assertEquals(t, "Type", "NS", res[0].Type)
	assertEqualsInt(t, "TTL", 3600, res[0].TtlInSec)
	assertEquals(t, "Target", "ns11.infomaniak.ch", res[0].Target)
}

func Test_GetDnsRecordsForZone_ParsesARecordCorrectly(t *testing.T) {
	client := aTestClient(`[{"id":5,"source":"subdomain","type":"A","ttl":60,"target":"1.1.1.1","updated_at":182637717,"dyndns_id":7}]`)

	res, _ := client.GetDnsRecordsForZone(context.TODO(), "example.com")

	assertEqualsInt(t, "ID", 5, res[0].ID)
	assertEquals(t, "Source", "subdomain", res[0].Source)
	assertEquals(t, "Type", "A", res[0].Type)
	assertEqualsInt(t, "TTL", 60, res[0].TtlInSec)
	assertEquals(t, "Target", "1.1.1.1", res[0].Target)
}

func Test_GetDnsRecordsForZone_ParsesTxtRecordCorrectly(t *testing.T) {
	client := aTestClient(`[{"id":35556917,"source":"alpha","type":"TXT","ttl":360,"target":"\"quotes \\\" backslashes \\\\000\"","updated_at":1445066462}]`)

	res, _ := client.GetDnsRecordsForZone(context.TODO(), "example.com")

	assertEqualsInt(t, "ID", 35556917, res[0].ID)
	assertEquals(t, "Source", "alpha", res[0].Source)
	assertEquals(t, "Type", "TXT", res[0].Type)
	assertEqualsInt(t, "TTL", 360, res[0].TtlInSec)
	assertEquals(t, "Target", `quotes " backslashes \000`, res[0].Target)
}

func Test_GetDnsRecordsForZone_ParsesCaaRecordCorrectly(t *testing.T) {
	client := aTestClient(`[{"id":450,"source":"libdns.test","type":"CAA","ttl":3600,"target":"1 issue \"127.0.0.1\"","updated_at":7,"description":{"flags":{"value":1},"tag":{"value":"issue"}}}]`)

	res, _ := client.GetDnsRecordsForZone(context.TODO(), "example.com")

	assertEqualsInt(t, "ID", 450, res[0].ID)
	assertEquals(t, "Source", "libdns.test", res[0].Source)
	assertEquals(t, "Type", "CAA", res[0].Type)
	assertEqualsInt(t, "TTL", 3600, res[0].TtlInSec)
	assertEquals(t, "Target", `1 issue "127.0.0.1"`, res[0].Target)
	assertEqualsInt(t, "Flags", 1, res[0].Description.Flags.Value)
	assertEquals(t, "Tag", "issue", res[0].Description.Tag.Value)
}

func Test_GetDnsRecordsForZone_ParsesCNameRecordCorrectly(t *testing.T) {
	client := aTestClient(`[{"id":33,"source":"test.libdns","type":"CNAME","ttl":3600,"target":"libdns.com","updated_at":5}]`)

	res, _ := client.GetDnsRecordsForZone(context.TODO(), "example.com")

	assertEqualsInt(t, "ID", 33, res[0].ID)
	assertEquals(t, "Source", "test.libdns", res[0].Source)
	assertEquals(t, "Type", "CNAME", res[0].Type)
	assertEqualsInt(t, "TTL", 3600, res[0].TtlInSec)
	assertEquals(t, "Target", `libdns.com`, res[0].Target)
}

func Test_GetDnsRecordsForZone_ParsesMxRecordCorrectly(t *testing.T) {
	client := aTestClient(`[{"id":778,"source":"libdns.test","type":"MX","ttl":3600,"target":"7 127.0.0.1","updated_at":9,"description":{"priority":{"value":7}}}]`)

	res, _ := client.GetDnsRecordsForZone(context.TODO(), "example.com")

	assertEqualsInt(t, "ID", 778, res[0].ID)
	assertEquals(t, "Source", "libdns.test", res[0].Source)
	assertEquals(t, "Type", "MX", res[0].Type)
	assertEqualsInt(t, "TTL", 3600, res[0].TtlInSec)
	assertEquals(t, "Target", `7 127.0.0.1`, res[0].Target)
	assertEqualsInt(t, "Priority", 7, res[0].Description.Priority.Value)
}

func Test_GetDnsRecordsForZone_ParsesSrvRecordCorrectly(t *testing.T) {
	client := aTestClient(`[{"id":73,"source":"libdns","type":"SRV","ttl":3600,"target":"10 0 5060 _sip._tcp.example.com","updated_at":7,"delegated_zone":{"id":8,"uri":"https:\/\/api.infomaniak.com\/2\/zones\/_tcp.example.com"},"description":{"priority":{"value":10},"port":{"value":5060},"weight":{"value":0},"protocol":{"value":"_tcp"}}}]`)

	res, _ := client.GetDnsRecordsForZone(context.TODO(), "example.com")

	assertEqualsInt(t, "ID", 73, res[0].ID)
	assertEquals(t, "Source", "libdns", res[0].Source)
	assertEquals(t, "Type", "SRV", res[0].Type)
	assertEqualsInt(t, "TTL", 3600, res[0].TtlInSec)
	assertEquals(t, "Target", `10 0 5060 _sip._tcp.example.com`, res[0].Target)
	assertEqualsInt(t, "Priority", 10, res[0].Description.Priority.Value)
	assertEqualsInt(t, "Weight", 0, res[0].Description.Weight.Value)
	assertEqualsInt(t, "Port", 5060, res[0].Description.Port.Value)
	assertEquals(t, "Protocol", "_tcp", res[0].Description.Protocol.Value)
}

func Test_GetDnsRecordsForZone_ReturnsErrorIfApiCallFails(t *testing.T) {
	client := aFailingTestClient(500, "")

	_, err := client.GetDnsRecordsForZone(context.TODO(), "subdomain.test.com")

	if err == nil {
		t.Fatalf("Expected error because API call failed")
	}
}

func Test_GetDnsRecordsForZone_ReturnsErrorIfApiReturnsError(t *testing.T) {
	client := aFailingTestClient(200, "some error message")

	_, err := client.GetDnsRecordsForZone(context.TODO(), "subdomain.test.com")

	if err == nil {
		t.Fatalf("Expected error because infomaniak API call returned error")
	}
}

func Test_DeleteRecord_CallsInfomaniakEndpointWithAuthHeader(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient("", &request)
	client.Token = "test-token"

	client.DeleteRecord(context.TODO(), "zone.com", "23")

	authHeader := request.Header.Get("Authorization")
	if authHeader != "Bearer test-token" {
		t.Fatalf("Authorization header not correct, expected: \"%s\", actual: \"%s\"", "Bearer test-token", authHeader)
	}
}

func Test_DeleteRecord_CallsInfomaniakEndpointWithContentTypeHeader(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient("", &request)

	client.DeleteRecord(context.TODO(), "zone.com", "23")

	contentTypeHeader := request.Header.Get("Content-Type")
	if contentTypeHeader != "application/json" {
		t.Fatalf("Content-Type header not correct, expected: \"%s\", actual: \"%s\"", "application/json", contentTypeHeader)
	}
}

func Test_DeleteRecord_CallsInfomaniakEndpointWithDeleteMethod(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient("", &request)

	client.DeleteRecord(context.TODO(), "zone.com", "23")

	if request.Method != http.MethodDelete {
		t.Fatalf("Wrong http method used, expected: \"%s\", actual: \"%s\"", http.MethodDelete, request.Method)
	}
}

func Test_DeleteRecord_CallsCorrectInfomaniakEndpoint(t *testing.T) {
	id := "333789"
	zone := "example.zone.com"
	expectedEndpoint := fmt.Sprintf("https://api.infomaniak.com/2/zones/%s/records/%s", zone, id)
	var request http.Request
	client := aRequestCapturingTestClient("", &request)

	client.DeleteRecord(context.TODO(), zone, id)

	endpoint := request.URL.String()
	if endpoint != expectedEndpoint {
		t.Fatalf("Wrong endpoint used, expected: \"%s\", actual: \"%s\"", expectedEndpoint, endpoint)
	}
}

func Test_DeleteRecord_ReturnsErrorIfApiCallFails(t *testing.T) {
	client := aFailingTestClient(500, "")

	err := client.DeleteRecord(context.TODO(), "subdomain.test.com", "25")

	if err == nil {
		t.Fatalf("Expected error because API call failed")
	}
}

func Test_DeleteRecord_ReturnsErrorIfApiReturnsError(t *testing.T) {
	client := aFailingTestClient(200, "some error message")

	err := client.DeleteRecord(context.TODO(), "subdomain.test.com", "83")

	if err == nil {
		t.Fatalf("Expected error because infomaniak API call returned error")
	}
}

func Test_CreateOrUpdateRecord_CallsInfomaniakEndpointWithAuthHeader(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient(`{"id": 5}`, &request)
	client.Token = "test-token"

	client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	authHeader := request.Header.Get("Authorization")
	if authHeader != "Bearer test-token" {
		t.Fatalf("Authorization header not correct, expected: \"%s\", actual: \"%s\"", "Bearer test-token", authHeader)
	}
}

func Test_CreateOrUpdateRecord_CallsInfomaniakEndpointWithContentTypeHeader(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient(`{"id": 5}`, &request)

	client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	contentTypeHeader := request.Header.Get("Content-Type")
	if contentTypeHeader != "application/json" {
		t.Fatalf("Content-Type header not correct, expected: \"%s\", actual: \"%s\"", "application/json", contentTypeHeader)
	}
}

func Test_CreateOrUpdateRecord_CallsInfomaniakEndpointWithPostMethodForNewRecords(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient(`{"id": 5}`, &request)

	client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{ID: 0})

	if request.Method != http.MethodPost {
		t.Fatalf("Wrong http method used, expected: \"%s\", actual: \"%s\"", http.MethodPost, request.Method)
	}
}

func Test_CreateOrUpdateRecord_CallsInfomaniakEndpointWithPutMethodForExistingRecords(t *testing.T) {
	var request http.Request
	client := aRequestCapturingTestClient(`{"id": 5}`, &request)

	client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{ID: 5})

	if request.Method != http.MethodPut {
		t.Fatalf("Wrong http method used, expected: \"%s\", actual: \"%s\"", http.MethodPut, request.Method)
	}
}

func Test_CreateOrUpdateRecord_CallsCorrectInfomaniakEndpointForNewRecords(t *testing.T) {
	zone := "sub.example.com"
	expectedEndpoint := fmt.Sprintf("https://api.infomaniak.com/2/zones/%s/records?with=records_description", zone)
	var request http.Request
	client := aRequestCapturingTestClient(`{"id": 5}`, &request)

	client.CreateOrUpdateRecord(context.TODO(), zone, IkRecord{ID: 0})

	endpoint := request.URL.String()
	if endpoint != expectedEndpoint {
		t.Fatalf("Wrong endpoint used, expected: \"%s\", actual: \"%s\"", expectedEndpoint, endpoint)
	}
}

func Test_CreateOrUpdateRecord_CallsCorrectInfomaniakEndpointForExistingRecords(t *testing.T) {
	id := 5
	zone := "sub.example.com"
	expectedEndpoint := fmt.Sprintf("https://api.infomaniak.com/2/zones/%s/records/%d?with=records_description", zone, id)
	var request http.Request
	client := aRequestCapturingTestClient(`{"id": 5}`, &request)

	client.CreateOrUpdateRecord(context.TODO(), zone, IkRecord{ID: id})

	endpoint := request.URL.String()
	if endpoint != expectedEndpoint {
		t.Fatalf("Wrong endpoint used, expected: \"%s\", actual: \"%s\"", expectedEndpoint, endpoint)
	}
}

func Test_CreateOrUpdateRecord_ReturnsUpdatedOrCreatedRecord(t *testing.T) {
	client := aTestClient(`{"id": 23}`)

	res, _ := client.CreateOrUpdateRecord(context.TODO(), "example.com", IkRecord{ID: 0})

	if res.ID != 23 {
		t.Fatalf("Expected created record with ID %d to be returned, got ID %d", 23, res.ID)
	}
}

func Test_CreateOrUpdateRecord_ParsesNSRecordCorrectly(t *testing.T) {
	client := aTestClient(`{"id":25,"source":".","type":"NS","ttl":3600,"target":"ns11.infomaniak.ch","updated_at":1659958248}`)

	res, _ := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	assertEqualsInt(t, "ID", 25, res.ID)
	assertEquals(t, "Source", ".", res.Source)
	assertEquals(t, "Type", "NS", res.Type)
	assertEqualsInt(t, "TTL", 3600, res.TtlInSec)
	assertEquals(t, "Target", "ns11.infomaniak.ch", res.Target)
}

func Test_CreateOrUpdateRecord_ParsesARecordCorrectly(t *testing.T) {
	client := aTestClient(`{"id":5,"source":"subdomain","type":"A","ttl":60,"target":"1.1.1.1","updated_at":182637717,"dyndns_id":7}`)

	res, _ := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	assertEqualsInt(t, "ID", 5, res.ID)
	assertEquals(t, "Source", "subdomain", res.Source)
	assertEquals(t, "Type", "A", res.Type)
	assertEqualsInt(t, "TTL", 60, res.TtlInSec)
	assertEquals(t, "Target", "1.1.1.1", res.Target)
}

func Test_CreateOrUpdateRecord_ParsesTxtRecordCorrectly(t *testing.T) {
	client := aTestClient(`{"id":35556917,"source":"alpha","type":"TXT","ttl":360,"target":"\"quotes \\\" backslashes \\\\000\"","updated_at":1445066462}`)

	res, _ := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	assertEqualsInt(t, "ID", 35556917, res.ID)
	assertEquals(t, "Source", "alpha", res.Source)
	assertEquals(t, "Type", "TXT", res.Type)
	assertEqualsInt(t, "TTL", 360, res.TtlInSec)
	assertEquals(t, "Target", `quotes " backslashes \000`, res.Target)
}

func Test_CreateOrUpdateRecord_ParsesCaaRecordCorrectly(t *testing.T) {
	client := aTestClient(`{"id":450,"source":"libdns.test","type":"CAA","ttl":3600,"target":"1 issue \"127.0.0.1\"","updated_at":7,"description":{"flags":{"value":1},"tag":{"value":"issue"}}}`)

	res, _ := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	assertEqualsInt(t, "ID", 450, res.ID)
	assertEquals(t, "Source", "libdns.test", res.Source)
	assertEquals(t, "Type", "CAA", res.Type)
	assertEqualsInt(t, "TTL", 3600, res.TtlInSec)
	assertEquals(t, "Target", `1 issue "127.0.0.1"`, res.Target)
	assertEqualsInt(t, "Flags", 1, res.Description.Flags.Value)
	assertEquals(t, "Tag", "issue", res.Description.Tag.Value)
}

func Test_CreateOrUpdateRecord_ParsesCNameRecordCorrectly(t *testing.T) {
	client := aTestClient(`{"id":33,"source":"test.libdns","type":"CNAME","ttl":3600,"target":"libdns.com","updated_at":5}`)

	res, _ := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	assertEqualsInt(t, "ID", 33, res.ID)
	assertEquals(t, "Source", "test.libdns", res.Source)
	assertEquals(t, "Type", "CNAME", res.Type)
	assertEqualsInt(t, "TTL", 3600, res.TtlInSec)
	assertEquals(t, "Target", `libdns.com`, res.Target)
}

func Test_CreateOrUpdateRecord_ParsesMxRecordCorrectly(t *testing.T) {
	client := aTestClient(`{"id":778,"source":"libdns.test","type":"MX","ttl":3600,"target":"7 127.0.0.1","updated_at":9,"description":{"priority":{"value":7}}}`)

	res, _ := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	assertEqualsInt(t, "ID", 778, res.ID)
	assertEquals(t, "Source", "libdns.test", res.Source)
	assertEquals(t, "Type", "MX", res.Type)
	assertEqualsInt(t, "TTL", 3600, res.TtlInSec)
	assertEquals(t, "Target", `7 127.0.0.1`, res.Target)
	assertEqualsInt(t, "Priority", 7, res.Description.Priority.Value)
}

func Test_CreateOrUpdateRecord_ParsesSrvRecordCorrectly(t *testing.T) {
	client := aTestClient(`{"id":73,"source":"libdns","type":"SRV","ttl":3600,"target":"10 0 5060 _sip._tcp.example.com","updated_at":7,"delegated_zone":{"id":8,"uri":"https:\/\/api.infomaniak.com\/2\/zones\/_tcp.example.com"},"description":{"priority":{"value":10},"port":{"value":5060},"weight":{"value":0},"protocol":{"value":"_tcp"}}}`)

	res, _ := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	assertEqualsInt(t, "ID", 73, res.ID)
	assertEquals(t, "Source", "libdns", res.Source)
	assertEquals(t, "Type", "SRV", res.Type)
	assertEqualsInt(t, "TTL", 3600, res.TtlInSec)
	assertEquals(t, "Target", `10 0 5060 _sip._tcp.example.com`, res.Target)
	assertEqualsInt(t, "Priority", 10, res.Description.Priority.Value)
	assertEqualsInt(t, "Weight", 0, res.Description.Weight.Value)
	assertEqualsInt(t, "Port", 5060, res.Description.Port.Value)
	assertEquals(t, "Protocol", "_tcp", res.Description.Protocol.Value)
}

func Test_CreateOrUpdateRecord_ReturnsErrorIfApiCallFails(t *testing.T) {
	client := aFailingTestClient(500, "")

	_, err := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	if err == nil {
		t.Fatalf("Expected error because API call failed")
	}
}

func Test_CreateorUpdateRecord_ReturnsErrorIfApiReturnsError(t *testing.T) {
	client := aFailingTestClient(200, "some error message")

	_, err := client.CreateOrUpdateRecord(context.TODO(), "test.com", IkRecord{})

	if err == nil {
		t.Fatalf("Expected error because infomaniak API call returned error")
	}
}
