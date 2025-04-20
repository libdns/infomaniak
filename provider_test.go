package infomaniak

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/libdns/libdns"
)

// TestClient instance of IkClient used to mock API calls
type TestClient struct {
	getter     func(ctx context.Context, zone string) ([]IkRecord, error)
	setter     func(ctx context.Context, zone string, record IkRecord) (*IkRecord, error)
	deleter    func(ctx context.Context, zone string, id string) error
	zoneGetter func(ctx context.Context, domain string) (string, error)
}

// GetDnsRecordsForZone implementation to fulfill IkClient interface
func (c *TestClient) GetDnsRecordsForZone(ctx context.Context, zone string) ([]IkRecord, error) {
	return c.getter(ctx, zone)
}

// CreateOrUpdateRecord implementation to fulfill IkClient interface
func (c *TestClient) CreateOrUpdateRecord(ctx context.Context, zone string, record IkRecord) (*IkRecord, error) {
	return c.setter(ctx, zone, record)
}

// DeleteRecord implementation to fulfill IkClient interface
func (c *TestClient) DeleteRecord(ctx context.Context, zone string, id string) error {
	return c.deleter(ctx, zone, id)
}

// GetFqdnOfZoneForDomain implementation to fulfill IkClient interface
func (c *TestClient) GetFqdnOfZoneForDomain(ctx context.Context, domain string) (string, error) {
	return c.zoneGetter(ctx, domain)
}

// assertEquals helper function that throws an error if the actual string value is not the expected value
func assertEquals(t *testing.T, name string, expected string, actual string) {
	if expected != actual {
		t.Fatalf("Expected %s \"%s\", got \"%s\"", name, expected, actual)
	}
}

// assertEqualsInt helper function that throws an error if the actual int value is not the expected value
func assertEqualsInt(t *testing.T, name string, expected int, actual int) {
	assertEquals(t, name, strconv.Itoa(expected), strconv.Itoa(actual))
}

func Test_GetZoneMapping_RemovesTrailingDotFromLibDnsZoneBeforeCallingClient(t *testing.T) {
	var zone string
	client := TestClient{zoneGetter: func(ctx context.Context, z string) (string, error) {
		zone = z
		return z, nil
	}}
	provider := Provider{client: &client}

	provider.getZoneMapping(context.TODO(), "zone.example.com.")

	assertEquals(t, "Zone", "zone.example.com", zone)
}

func Test_GetZoneMapping_WorksIfProvidedZoneHasNoTrailingDot(t *testing.T) {
	var zone string
	client := TestClient{zoneGetter: func(ctx context.Context, argZone string) (string, error) {
		zone = argZone
		return argZone, nil
	}}
	provider := Provider{client: &client}

	provider.getZoneMapping(context.TODO(), "zone.example.com")

	assertEquals(t, "Zone", "zone.example.com", zone)
}

func Test_GetZoneMapping_ReturnsLibDnsZoneWithoutTrailingDot(t *testing.T) {
	client := TestClient{zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil }}
	provider := Provider{client: &client}

	res, _ := provider.getZoneMapping(context.TODO(), "zone.example.com.")

	assertEquals(t, "Zone", "zone.example.com", res.LibDnsZone)
}

func Test_GetZoneMapping_ReturnsInfomaniakManagedZone(t *testing.T) {
	client := TestClient{zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "zone.managed.com", nil }}
	provider := Provider{client: &client}

	res, _ := provider.getZoneMapping(context.TODO(), "zone.example.com.")

	assertEquals(t, "Zone", "zone.managed.com", res.InfomaniakManagedZone)
}

func Test_GetRecords_PassesZoneManagedByInfomaniakToClient(t *testing.T) {
	var zone string
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			zone = argZone
			return []IkRecord{}, nil
		},
	}
	provider := Provider{client: &client}

	provider.GetRecords(context.TODO(), "zone.example.com")

	assertEquals(t, "Zone", "example.com", zone)
}

func Test_GetRecords_ReturnsRecord(t *testing.T) {
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			return []IkRecord{IkRecord{Source: "zone", Type: "libdns_infomaniak_test"}}, nil
		},
	}
	provider := Provider{client: &client}

	res, _ := provider.GetRecords(context.TODO(), "zone.example.com")
	if len(res) == 0 {
		t.Fatalf("Did not get any record")
	}

	if res[0].(libdns.RR).Type != "libdns_infomaniak_test" {
		t.Fatalf("Did not get expected record")
	}
}

func Test_GetRecords_DoesNotReturnRecordInParentZone(t *testing.T) {
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			return []IkRecord{IkRecord{Source: "@"}}, nil
		},
	}
	provider := Provider{client: &client}

	res, _ := provider.GetRecords(context.TODO(), "zone.example.com")
	if len(res) != 0 {
		t.Fatalf("Got record but did not expect any")
	}
}

func Test_AppendRecords_PassesZoneManagedByInfomaniakToClient(t *testing.T) {
	var zone string
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) {
			zone = argZone
			return &record, nil
		},
	}
	provider := Provider{client: &client}

	provider.AppendRecords(context.TODO(), "zone.example.com", []libdns.Record{libdns.Address{}})

	assertEquals(t, "Zone", "example.com", zone)
}

func Test_AppendRecords_PassesRecordToClient(t *testing.T) {
	var passedRecord IkRecord
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) {
			passedRecord = record
			return &record, nil
		},
	}
	provider := Provider{client: &client}

	provider.AppendRecords(context.TODO(), "zone.example.com", []libdns.Record{libdns.RR{Type: "libdns_infomaniak_test"}})

	if passedRecord.Type != "libdns_infomaniak_test" {
		t.Fatalf("Did not pass expected record")
	}
}

func Test_AppendRecords_ReturnsCreatedRecord(t *testing.T) {
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) {
			return &IkRecord{Type: "libdns_infomaniak_test"}, nil
		},
	}
	provider := Provider{client: &client}

	res, _ := provider.AppendRecords(context.TODO(), "zone.example.com", []libdns.Record{libdns.RR{}})

	if len(res) == 0 {
		t.Fatalf("Did not get any record")
	}

	if res[0].(libdns.RR).Type != "libdns_infomaniak_test" {
		t.Fatalf("Did not get expected record")
	}
}

func Test_SetRecords_PassesZoneManagedByInfomaniakToClient(t *testing.T) {
	var zone string
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{}, nil },
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) {
			zone = argZone
			return &record, nil
		},
	}
	provider := Provider{client: &client}

	provider.SetRecords(context.TODO(), "zone.example.com", []libdns.Record{libdns.Address{}})

	assertEquals(t, "Zone", "example.com", zone)
}

func Test_SetRecords_PassesRecordToClient(t *testing.T) {
	var passedRecord IkRecord
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{}, nil },
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) {
			passedRecord = record
			return &record, nil
		},
	}
	provider := Provider{client: &client}

	provider.SetRecords(context.TODO(), "zone.example.com", []libdns.Record{libdns.RR{Type: "libdns_infomaniak_test"}})

	if passedRecord.Type != "libdns_infomaniak_test" {
		t.Fatalf("Did not pass expected record")
	}
}

func Test_SetRecords_ReturnsCreatedRecord(t *testing.T) {
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{}, nil },
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) {
			return &IkRecord{Type: "libdns_infomaniak_test"}, nil
		},
	}
	provider := Provider{client: &client}

	res, _ := provider.SetRecords(context.TODO(), "zone.example.com", []libdns.Record{libdns.RR{}})

	if len(res) == 0 {
		t.Fatalf("Did not get any record")
	}

	if res[0].(libdns.RR).Type != "libdns_infomaniak_test" {
		t.Fatalf("Did not get expected record")
	}
}

func Test_SetRecords_DeletesRecordWithSameTypeAndSource(t *testing.T) {
	existingRec := IkRecord{Type: "type", Source: "sub"}
	newRec := libdns.RR{Type: "type", Name: "sub.test"}

	deleteCalled := false
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			return []IkRecord{existingRec}, nil
		},
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) { return &record, nil },
		deleter: func(ctx context.Context, zone, id string) error {
			deleteCalled = true
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.SetRecords(context.TODO(), "test.example.com", []libdns.Record{newRec})

	if deleteCalled {
		t.Fatalf("Expected existing record to be deleted, but was not")
	}
}

func Test_SetRecords_DeletesAlreadyExistingRecordsOnlyOnce(t *testing.T) {
	existingRec := IkRecord{Type: "test_type", Source: "sub"}
	newRec1 := libdns.RR{Type: "test_type", Name: "sub"}
	newRec2 := libdns.RR{Type: "test_type", Name: "sub"}

	deleteCalled := 0
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			return []IkRecord{existingRec}, nil
		},
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) { return &record, nil },
		deleter: func(ctx context.Context, zone, id string) error {
			deleteCalled = deleteCalled + 1
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.SetRecords(context.TODO(), "example.com", []libdns.Record{newRec1, newRec2})

	if deleteCalled != 1 {
		t.Fatalf("Expected existing record to be deleted once, delete was called %d times", deleteCalled)
	}
}

func Test_SetRecords_DoesNotDeleteExistingRecordOfDifferentType(t *testing.T) {
	existingRec := IkRecord{Type: "type1", Source: "sub"}
	newRec := libdns.RR{Type: "type2", Name: "sub"}

	deleteCalled := false
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			return []IkRecord{existingRec}, nil
		},
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) { return &record, nil },
		deleter: func(ctx context.Context, zone, id string) error {
			deleteCalled = true
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.SetRecords(context.TODO(), "example.com", []libdns.Record{newRec})

	if deleteCalled {
		t.Fatalf("Expected no record to be deleted")
	}
}

func Test_SetRecords_DoesNotDeleteExistingRecordWithDifferentSource(t *testing.T) {
	existingRec := IkRecord{Type: "type", Source: "sub1"}
	newRec := libdns.RR{Type: "type", Name: "sub2"}

	deleteCalled := false
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			return []IkRecord{existingRec}, nil
		},
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) { return &record, nil },
		deleter: func(ctx context.Context, zone, id string) error {
			deleteCalled = true
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.SetRecords(context.TODO(), "example.com", []libdns.Record{newRec})

	if deleteCalled {
		t.Fatalf("Expected no record to be deleted")
	}
}

func Test_SetRecords_DoesNotDeleteExistingRecordInParentZone(t *testing.T) {
	existingRec := IkRecord{Type: "type", Source: "@"}
	newRec := libdns.RR{Type: "type", Name: "@"}

	deleteCalled := false
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			return []IkRecord{existingRec}, nil
		},
		setter: func(ctx context.Context, argZone string, record IkRecord) (*IkRecord, error) { return &record, nil },
		deleter: func(ctx context.Context, zone, id string) error {
			deleteCalled = true
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.SetRecords(context.TODO(), "subzone.example.com", []libdns.Record{newRec})

	if deleteCalled {
		t.Fatalf("Expected no record to be deleted")
	}
}

func Test_DeleteRecords_LoadsExistingRecordsWithInfomaniakManagedZone(t *testing.T) {
	var zone string
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			zone = argZone
			return []IkRecord{}, nil
		},
		deleter: func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	provider.DeleteRecords(context.TODO(), "zone.example.com", []libdns.Record{libdns.Address{}})

	assertEquals(t, "Zone", "example.com", zone)
}

func Test_DeleteRecords_DeletesExistingRecordWithInfomaniakManagedZone(t *testing.T) {
	existingRec := IkRecord{Source: "sub.zone"}
	recToDelete := libdns.RR{Name: "sub"}

	var zone string
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter: func(ctx context.Context, argZone, id string) error {
			zone = argZone
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.DeleteRecords(context.TODO(), "zone.example.com", []libdns.Record{recToDelete})

	assertEquals(t, "Zone", "example.com", zone)
}

func Test_DeleteRecords_DeletesMatchingExistingRecord(t *testing.T) {
	existingRec := IkRecord{ID: 1893, Source: "sub1"}
	nonMatchingExistingRec := IkRecord{ID: 23, Source: "sub2"}
	recToDelete := libdns.RR{Name: "sub1"}

	deletedIds := []string{}
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter: func(ctx context.Context, argZone string) ([]IkRecord, error) {
			return []IkRecord{existingRec, nonMatchingExistingRec}, nil
		},
		deleter: func(ctx context.Context, argZone, id string) error {
			deletedIds = append(deletedIds, id)
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete})

	if len(deletedIds) != 1 {
		t.Fatalf("Nothing was deleted")
	}

	if deletedIds[0] != "1893" {
		t.Fatalf("Expected ID %s to be deleted, got %s", "1893", deletedIds[0])
	}
}

func Test_DeleteRecords_DoesNotTryToDeleteNonExistingRecord(t *testing.T) {
	deleteCalled := false
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{}, nil },
		deleter: func(ctx context.Context, argZone, id string) error {
			deleteCalled = true
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{libdns.RR{Name: "sub1"}})

	if deleteCalled {
		t.Fatalf("Delete was called although not expected")
	}
}

func Test_DeleteRecords_DeletesExistingRecordOnlyOnce(t *testing.T) {
	existingRec := IkRecord{ID: 1893, Source: "sub1"}
	recToDelete := libdns.RR{Name: "sub1"}
	recToDelete2 := libdns.RR{Name: "sub1"}

	deleteCallCount := 0
	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter: func(ctx context.Context, argZone, id string) error {
			deleteCallCount = deleteCallCount + 1
			return nil
		},
	}
	provider := Provider{client: &client}

	provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete, recToDelete2})

	if deleteCallCount != 1 {
		t.Fatalf("Expected for delete only to be called once, was called %d", deleteCallCount)
	}
}

func Test_DeleteRecords_ReturnsDeletedRecord(t *testing.T) {
	existingRec := IkRecord{Source: "sub1", Type: "test_type"}
	recToDelete := libdns.RR{Name: "sub1"}

	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter:    func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	res, _ := provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete})

	if len(res) != 1 {
		t.Fatalf("Expected 1 deleted record, got %d", len(res))
	}

	if res[0].(libdns.RR).Type != "test_type" {
		t.Fatalf("Method did not return deleted record")
	}
}

func Test_DeleteRecords_DoesNotDeleteRecordOfParentZone(t *testing.T) {
	existingRec := IkRecord{Source: "sub1", Target: "127.0.0.1"}
	recToDelete := libdns.RR{Name: "sub1", Data: "127.0.0.1"}

	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return "example.com", nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter:    func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	res, _ := provider.DeleteRecords(context.TODO(), "test.example.com", []libdns.Record{recToDelete})

	if len(res) != 0 {
		t.Fatalf("Expected no record to be deleted")
	}
}

func Test_DeleteRecords_DeletesRecordWhenSourceAndTtlMatches(t *testing.T) {
	existingRec := IkRecord{Source: "sub1", TtlInSec: 60}
	recToDelete := libdns.RR{Name: "sub1", TTL: time.Duration(60 * time.Second)}

	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter:    func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	res, _ := provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete})

	if len(res) != 1 {
		t.Fatalf("Expected record to be deleted, but it wasn't")
	}
}

func Test_DeleteRecords_DoesNotDeleteRecordIfTtlIsNotMatching(t *testing.T) {
	existingRec := IkRecord{Source: "sub1", TtlInSec: 60}
	recToDelete := libdns.RR{Name: "sub1", TTL: time.Duration(61)}

	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter:    func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	res, _ := provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete})

	if len(res) != 0 {
		t.Fatalf("Expected no record to be deleted")
	}
}

func Test_DeleteRecords_DeletesRecordWhenSourceAndTypeMatches(t *testing.T) {
	existingRec := IkRecord{Source: "sub1", Type: "TXT"}
	recToDelete := libdns.RR{Name: "sub1", Type: "TXT"}

	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter:    func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	res, _ := provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete})

	if len(res) != 1 {
		t.Fatalf("Expected record to be deleted, but it wasn't")
	}
}

func Test_DeleteRecords_DoesNotDeleteRecordWhenTypeDoesNotMatch(t *testing.T) {
	existingRec := IkRecord{Source: "sub1", Type: "TXT"}
	recToDelete := libdns.RR{Name: "sub1", Type: "TXT-2"}

	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter:    func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	res, _ := provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete})

	if len(res) != 0 {
		t.Fatalf("Expected no record to be deleted")
	}
}

func Test_DeleteRecords_DeletesRecordWhenSourceAndTarget(t *testing.T) {
	existingRec := IkRecord{Source: "sub1", Target: "127.0.0.1"}
	recToDelete := libdns.RR{Name: "sub1", Data: "127.0.0.1"}

	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter:    func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	res, _ := provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete})

	if len(res) != 1 {
		t.Fatalf("Expected record to be deleted, but it wasn't")
	}
}

func Test_DeleteRecords_DoesNotDeleteRecordIfTargetDoesNotMatch(t *testing.T) {
	existingRec := IkRecord{Source: "sub1", Target: "127.0.0.1"}
	recToDelete := libdns.RR{Name: "sub1", Data: "127.0.0.2"}

	client := TestClient{
		zoneGetter: func(ctx context.Context, argZone string) (string, error) { return argZone, nil },
		getter:     func(ctx context.Context, argZone string) ([]IkRecord, error) { return []IkRecord{existingRec}, nil },
		deleter:    func(ctx context.Context, argZone, id string) error { return nil },
	}
	provider := Provider{client: &client}

	res, _ := provider.DeleteRecords(context.TODO(), "example.com", []libdns.Record{recToDelete})

	if len(res) != 0 {
		t.Fatalf("Expected no record to be deleted")
	}
}
