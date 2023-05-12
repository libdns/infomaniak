package infomaniak

import (
	"testing"

	"github.com/libdns/libdns"
)

func Test_ToLibDnsRecord_MapsAllProperties(t *testing.T) {
	ikRec := IkRecord{
		ID:       "123456",
		Type:     "MX",
		Target:   "127.0.0.1",
		TtlInSec: 3600,
		Priority: 17,
	}

	libRec := ikRec.ToLibDnsRecord("")
	assertEquals(t, "ID", ikRec.ID, libRec.ID)
	assertEquals(t, "Type", ikRec.Type, libRec.Type)
	assertEquals(t, "Value", ikRec.Target, libRec.Value)
	assertEqualsInt(t, "TTL", int(ikRec.TtlInSec), int(int64(libRec.TTL)))
	assertEqualsInt(t, "Priority", int(ikRec.Priority), libRec.Priority)
}

func Test_ToLibDnsRecord_ReturnsRelativeName(t *testing.T) {
	subzone := "sub"
	zone := "example.com"

	ikRec := IkRecord{
		SourceIdn: subzone + "." + zone,
	}

	libRec := ikRec.ToLibDnsRecord(zone)
	assertEquals(t, "Name", subzone, libRec.Name)
}

func Test_ToInfomaniakRecord_MapsAllProperties(t *testing.T) {
	libRec := libdns.Record{
		ID:       "123456",
		Type:     "MX",
		Value:    "127.0.0.1",
		TTL:      3600,
		Priority: 17,
	}

	ikRec := ToInfomaniakRecord(&libRec, "")
	assertEquals(t, "ID", libRec.ID, ikRec.ID)
	assertEquals(t, "Type", libRec.Type, ikRec.Type)
	assertEquals(t, "Value", libRec.Value, ikRec.Target)
	assertEqualsInt(t, "TTL", int(libRec.TTL), int(ikRec.TtlInSec))
	assertEqualsInt(t, "Priority", libRec.Priority, int(ikRec.Priority))
}

func Test_ToInfomaniakRecord_DefaultTtlIsAppliedIfNoTtlProvided(t *testing.T) {
	ikRec := ToInfomaniakRecord(&libdns.Record{}, "")
	assertEqualsInt(t, "TTL", defaultTtlSecs, int(ikRec.TtlInSec))
}

func Test_ToInfomaniakRecord_DefaultPriorityIsAppliedIfPriority(t *testing.T) {
	ikRec := ToInfomaniakRecord(&libdns.Record{}, "")
	assertEqualsInt(t, "Priority", defaultPriority, int(ikRec.Priority))
}

func Test_ToInfomaniakRecord_SetsAbsoluteNameToSource(t *testing.T) {
	subzone := "sub"
	zone := "example.com"
	ikRec := ToInfomaniakRecord(&libdns.Record{Name: subzone}, zone)
	assertEquals(t, "SourceIdn", subzone+"."+zone, ikRec.SourceIdn)
}
