package testing

import (
	"context"
	"testing"
	"time"

	"github.com/libdns/infomaniak"
	"github.com/libdns/libdns"
)

// Put your API token here - do not forget to remove it before committing!
const apiToken = "<YOUR_TOKEN>"

// Use a subdomain that you normally don't use.
// For example use "test.example.com." to prevent that your actual dns records are changed.
// Make sure you append a trailing "." to your domain as in the example above
const zone = "<YOUR_(SUB)_DOMAIN>"

// Provider used for integration test
var provider = infomaniak.Provider{APIToken: apiToken}

// contains the the created test records to clean up after the test
var testRecords = make([]libdns.Record, 0)

// cleanup ensures that all created records are removed after each test
func cleanup() {
	provider.DeleteRecords(context.TODO(), zone, testRecords)
	testRecords = make([]libdns.Record, 0)
}

// appendRecord calls provider, handles error and ensures that the appended records will be deleted at the end of the test
func appendRecord(t *testing.T, rec libdns.Record) []libdns.Record {
	appendedRecords, err := provider.AppendRecords(context.TODO(), zone, []libdns.Record{rec})
	if err != nil {
		t.Fatal(err)
	}
	testRecords = append(testRecords, appendedRecords...)
	return appendedRecords
}

// setRecord calls provider, handles error and ensures that the set records will be deleted at the end of the test
func setRecord(t *testing.T, rec libdns.Record) []libdns.Record {
	return setRecordInSpecificZone(t, zone, rec)
}

// setRecordInSpecificZone calls provider for another than the default zone,
// handles error and ensures that the set records will be deleted at the end of the test
func setRecordInSpecificZone(t *testing.T, specificZone string, rec libdns.Record) []libdns.Record {
	setRecords, err := provider.SetRecords(context.TODO(), specificZone, []libdns.Record{rec})
	if err != nil {
		t.Fatal(err)
	}
	testRecords = append(testRecords, setRecords...)
	return setRecords
}

// deleteRecord calls provider, handles error and ensures that the deleted records will not be deleted again at the end of the test
func deleteRecord(t *testing.T, rec libdns.Record) []libdns.Record {
	deletedRecs, err := provider.DeleteRecords(context.TODO(), zone, []libdns.Record{rec})
	if err != nil {
		t.Fatal(err)
	}

	indexOfDeletedRec := -1
	for i, testRec := range testRecords {
		if testRec == rec {
			indexOfDeletedRec = i
			break
		}
	}

	if len(testRecords) <= 1 {
		testRecords = make([]libdns.Record, 0)
	} else if indexOfDeletedRec > -1 {
		testRecords[indexOfDeletedRec] = testRecords[len(testRecords)-1]
		testRecords = testRecords[:len(testRecords)-1]
	}
	return deletedRecs
}

// getRecords calls provider, handles error and returns records that exist for zone
func getRecords(t *testing.T, zone string) []libdns.Record {
	result, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// aTestRecord returns a record that can be used for testing purposes
func aTestRecord(name string, recType string, value string) libdns.Record {
	return libdns.RR{
		Type: recType,
		Name: libdns.RelativeName(name, zone),
		Data: value,
		TTL:  time.Duration(3600 * time.Second),
	}
}

// assertExists ensures that a record exists based on it's name, type and value
func assertExists(t *testing.T, record libdns.Record) {
	if !isRecordExisting(t, record) {
		t.Fatalf("Expected for record %#v to exist, but it does not", record)
	}
}

// assertNotExists ensures that a record with the given name, type and value does not exist
func assertNotExists(t *testing.T, record libdns.Record) {
	if isRecordExisting(t, record) {
		t.Fatalf("Expected for record %#v to not exist, but it does", record)
	}
}

// isRecordExisting returns if a record with given name, type and value exists
func isRecordExisting(t *testing.T, record libdns.Record) bool {
	rr := record.RR()
	existingRecs := getRecords(t, zone)
	for _, existingRec := range existingRecs {
		existingRr := existingRec.RR()
		if existingRr.Name == rr.Name && existingRr.Type == rr.Type && existingRr.Data == rr.Data {
			return true
		}
	}
	return false
}

func Test_DeleteRecords_DeletesRecordByNameAndType(t *testing.T) {
	defer cleanup()

	recToDeleteWithoutId := aTestRecord(zone, "A", "127.0.0.1")
	setRecord(t, recToDeleteWithoutId)
	deleteRecord(t, recToDeleteWithoutId)
	assertNotExists(t, recToDeleteWithoutId)
}

func Test_AppendRecords_AppendsNewRecord(t *testing.T) {
	defer cleanup()

	recToAppend := aTestRecord(zone, "A", "127.0.0.1")
	appendedRecords := appendRecord(t, recToAppend)
	if len(appendedRecords) != 1 {
		t.Fatalf("Expected 1 record appended, got %d", len(appendedRecords))
	}
	assertExists(t, appendedRecords[0])
}

func Test_AppendRecords_DoesNotOverwriteExistingRecordWithSameNameAndType(t *testing.T) {
	defer cleanup()

	originalRecord := aTestRecord(zone, "A", "127.0.0.1")
	appendRecord(t, originalRecord)

	recThatShouldNotOverwriteFirst := originalRecord.RR()
	recThatShouldNotOverwriteFirst.Data = "127.0.0.0"
	addedRecords, err := provider.AppendRecords(context.TODO(), zone, []libdns.Record{recThatShouldNotOverwriteFirst})
	if err == nil {
		testRecords = append(testRecords, addedRecords...)
	}

	assertExists(t, originalRecord)
}

func Test_SetRecords_CreatesNewRecord(t *testing.T) {
	defer cleanup()

	recToCreate := aTestRecord(zone, "TXT", "127.0.0.1")
	setRecs := setRecord(t, recToCreate)

	if len(setRecs) != 1 {
		t.Fatalf("Expected 1 record updated, got %d", len(setRecs))
	}
	assertExists(t, setRecs[0])
}

func Test_SetRecords_OverwritesExistingRecordWithSameNameAndType(t *testing.T) {
	defer cleanup()

	recToUpdate := aTestRecord(zone, "MX", "3 127.0.0.1")
	updatedRec := setRecord(t, recToUpdate)[0].RR()

	updatedRec.Data = "7 127.0.0.0"
	result := setRecord(t, updatedRec)

	if len(result) != 1 {
		t.Fatalf("Expected 1 record updated, got %d", len(result))
	}
	assertNotExists(t, recToUpdate)
	assertExists(t, updatedRec)
}

func Test_GetRecords_DoesNotReturnRecordsOfParentZone(t *testing.T) {
	defer cleanup()

	setRecord(t, aTestRecord(zone, "MX", "8 127.0.0.1"))
	result := getRecords(t, "subzone."+zone)
	if len(result) > 0 {
		t.Fatalf("Expected 0 records, got %d", len(result))
	}
}

func Test_GetRecords_ReturnsRecordOfChildZone(t *testing.T) {
	defer cleanup()

	setRecord(t, aTestRecord("subzone."+zone, "MX", "7 127.0.0.1"))
	result := getRecords(t, zone)
	if len(result) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result))
	}
}
