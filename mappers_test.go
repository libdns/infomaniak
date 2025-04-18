package infomaniak

import (
	"testing"

	"fmt"

	"github.com/libdns/libdns"
)

func Test_ToLibDnsRecord_MapsAllProperties(t *testing.T) {
	ikRec := IkRecord{
		ID:       123456,
		Type:     "MX",
		Target:   "127.0.0.1",
		TtlInSec: 3600,
	}

	libRec := ikRec.ToLibDnsRecord()
	assertEquals(t, "ID", fmt.Sprint(ikRec.ID), libRec.ID)
	assertEquals(t, "Type", ikRec.Type, libRec.Type)
	assertEquals(t, "Value", ikRec.Target, libRec.Value)
	assertEqualsInt(t, "TTL", ikRec.TtlInSec, 3600)
	assertEqualsInt(t, "Priority", 10, int(libRec.Priority))
}

func Test_ToInfomaniakRecord_MapsAllProperties(t *testing.T) {
	libRec := libdns.Record{
		ID:       "123456",
		Type:     "MX",
		Value:    "127.0.0.1",
		TTL:      3600,
		Priority: 17,
	}

	ikRec := ToInfomaniakRecord(&libRec)
	assertEquals(t, "ID", libRec.ID, fmt.Sprint(ikRec.ID))
	assertEquals(t, "Type", libRec.Type, ikRec.Type)
	assertEquals(t, "Value", libRec.Value, ikRec.Target)
	assertEqualsInt(t, "TTL", int(libRec.TTL), ikRec.TtlInSec)
}

func Test_ToInfomaniakRecord_DefaultTtlIsAppliedIfNoTtlProvided(t *testing.T) {
	ikRec := ToInfomaniakRecord(&libdns.Record{})
	assertEqualsInt(t, "TTL", defaultTtlSecs, int(ikRec.TtlInSec))
}
