package infomaniak

import (
	"testing"
	"time"

	"github.com/libdns/libdns"
)

func Test_ToLibDnsRecord_CreatesAddressRecordForARecord(t *testing.T) {
	ikRec := IkRecord{
		Source:   "test.zone",
		Type:     "A",
		Target:   "127.0.0.1",
		TtlInSec: 3600,
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "example.com", LibDnsZone: "zone.example.com"})
	addressRec, isTypeOk := libRec.(libdns.Address)
	if !isTypeOk {
		t.Fatalf("Expected libdns.Address type")
	}
	assertEquals(t, "Name", "test", addressRec.Name)
	assertEquals(t, "IP", "127.0.0.1", addressRec.IP.String())
	assertEqualsInt(t, "TTL", 3600, int(addressRec.TTL.Seconds()))
}

func Test_ToLibDnsRecord_FailsToCreateAddressRecordForInvalidIP(t *testing.T) {
	ikRec := IkRecord{
		Target: "127-0-0-1",
		Type:   "A",
	}

	_, err := ikRec.ToLibDnsRecord(&ZoneMapping{})
	if err == nil {
		t.Fatalf("Expected error due to invalid IP")
	}
}

func Test_ToLibDnsRecord_CreatesAddressRecordForAAAARecord(t *testing.T) {
	ikRec := IkRecord{
		Source:   "test.zone",
		Type:     "AAAA",
		Target:   "::1",
		TtlInSec: 60,
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "example.com", LibDnsZone: "zone.example.com"})
	addressRec, isTypeOk := libRec.(libdns.Address)
	if !isTypeOk {
		t.Fatalf("Expected libdns.Address type")
	}
	assertEquals(t, "Name", "test", addressRec.Name)
	assertEquals(t, "IP", "::1", addressRec.IP.String())
	assertEqualsInt(t, "TTL", 60, int(addressRec.TTL.Seconds()))
}

func Test_ToLibDnsRecord_CreatesCaaRecord(t *testing.T) {
	ikRec := IkRecord{
		Source:      "subdomain.zone",
		Type:        "CAA",
		Target:      `1 issue "example.com"`,
		TtlInSec:    60,
		Description: IkRecordDescription{Tag: IkStringValueAttribute{Value: "issue"}, Flags: IkIntValueAttribute{Value: 1}},
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "example.com", LibDnsZone: "zone.example.com"})
	caaRec, isTypeOk := libRec.(libdns.CAA)
	if !isTypeOk {
		t.Fatalf("Expected libdns.CAA type")
	}
	assertEquals(t, "Name", "subdomain", caaRec.Name)
	assertEquals(t, "Tag", "issue", caaRec.Tag)
	assertEqualsInt(t, "Flags", 1, int(caaRec.Flags))
	assertEqualsInt(t, "TTL", 60, int(caaRec.TTL.Seconds()))
	assertEquals(t, "Value", "example.com", caaRec.Value)
}

func Test_ToLibDnsRecord_CreatesCNameRecord(t *testing.T) {
	ikRec := IkRecord{
		Source:   "subdomain.zone",
		Type:     "CNAME",
		Target:   "subdomain.alias.com",
		TtlInSec: 60,
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "example.com", LibDnsZone: "zone.example.com"})
	cnameRec, isTypeOk := libRec.(libdns.CNAME)
	if !isTypeOk {
		t.Fatalf("Expected libdns.CNAME type")
	}
	assertEquals(t, "Name", "subdomain", cnameRec.Name)
	assertEquals(t, "Target", "subdomain.alias.com", cnameRec.Target)
	assertEqualsInt(t, "TTL", 60, int(cnameRec.TTL.Seconds()))
}

func Test_ToLibDnsRecord_CreatesMxRecord(t *testing.T) {
	ikRec := IkRecord{
		Source:      "test.zone",
		Type:        "MX",
		Target:      "23 1.1.1.1",
		TtlInSec:    60,
		Description: IkRecordDescription{Priority: IkIntValueAttribute{Value: 23}},
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "example.com", LibDnsZone: "zone.example.com"})
	mxRec, isTypeOk := libRec.(libdns.MX)
	if !isTypeOk {
		t.Fatalf("Expected libdns.MX type")
	}
	assertEquals(t, "Name", "test", mxRec.Name)
	assertEquals(t, "Target", "1.1.1.1", mxRec.Target)
	assertEqualsInt(t, "TTL", 60, int(mxRec.TTL.Seconds()))
	assertEqualsInt(t, "Preference", 23, int(mxRec.Preference))
}

func Test_ToLibDnsRecord_CreatesNsRecord(t *testing.T) {
	ikRec := IkRecord{
		Source:   "test.zone",
		Type:     "NS",
		Target:   "ns11.infomaniak.ch",
		TtlInSec: 60,
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "example.com", LibDnsZone: "zone.example.com"})
	nsRec, isTypeOk := libRec.(libdns.NS)
	if !isTypeOk {
		t.Fatalf("Expected libdns.NS type")
	}
	assertEquals(t, "Name", "test", nsRec.Name)
	assertEquals(t, "Target", "ns11.infomaniak.ch", nsRec.Target)
	assertEqualsInt(t, "TTL", 60, int(nsRec.TTL.Seconds()))
}

func Test_ToLibDnsRecord_CreatesServiceRecord(t *testing.T) {
	ikRec := IkRecord{
		Source:      "_sip.test",
		Type:        "SRV",
		Target:      "10 0 5060 target.test.com",
		TtlInSec:    60,
		Description: IkRecordDescription{Port: IkIntValueAttribute{Value: 5060}, Priority: IkIntValueAttribute{Value: 10}, Weight: IkIntValueAttribute{Value: 0}, Protocol: IkStringValueAttribute{Value: "_tcp"}},
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "example.domain.com", LibDnsZone: "domain.com"})
	srvRec, isTypeOk := libRec.(libdns.SRV)
	if !isTypeOk {
		t.Fatalf("Expected libdns.SRV type")
	}
	assertEquals(t, "Service", "sip", srvRec.Service)
	assertEquals(t, "Transport", "tcp", srvRec.Transport)
	assertEquals(t, "Name", "_sip.test.example", srvRec.Name)
	assertEqualsInt(t, "TTL", 60, int(srvRec.TTL.Seconds()))
	assertEqualsInt(t, "Priority", 10, int(srvRec.Priority))
	assertEqualsInt(t, "Weight", 0, int(srvRec.Weight))
	assertEqualsInt(t, "Port", 5060, int(srvRec.Port))
	assertEquals(t, "Target", "target.test.com", srvRec.Target)
}

func Test_ToLibDnsRecord_CreatesTxtRecord(t *testing.T) {
	ikRec := IkRecord{
		Source:   "example",
		Type:     "TXT",
		Target:   "This is an awesome domain! Definitely not spammy.",
		TtlInSec: 60,
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "domain.com", LibDnsZone: "example.domain.com"})
	txtRec, isTypeOk := libRec.(libdns.TXT)
	if !isTypeOk {
		t.Fatalf("Expected libdns.TXT type")
	}
	assertEquals(t, "Name", "@", txtRec.Name)
	assertEquals(t, "Text", "This is an awesome domain! Definitely not spammy.", txtRec.Text)
	assertEqualsInt(t, "TTL", 60, int(txtRec.TTL.Seconds()))
}

func Test_ToLibDnsRecord_CreatesRrForNonTypedRecordType(t *testing.T) {
	ikRec := IkRecord{
		Source:   "test",
		Type:     "RNAME",
		Target:   "admin.example.com",
		TtlInSec: 60,
	}

	libRec, _ := ikRec.ToLibDnsRecord(&ZoneMapping{InfomaniakManagedZone: "domain.com", LibDnsZone: "test.domain.com"})
	rec, isTypeOk := libRec.(libdns.RR)
	if !isTypeOk {
		t.Fatalf("Expected libdns.RR type")
	}
	assertEquals(t, "Name", "@", rec.Name)
	assertEquals(t, "Type", "RNAME", rec.Type)
	assertEquals(t, "Text", "admin.example.com", rec.Data)
	assertEqualsInt(t, "TTL", 60, int(rec.TTL.Seconds()))
}

func Test_ToInfomaniakRecord_MapsAllProperties(t *testing.T) {
	libRec := libdns.RR{
		Name: "@",
		Type: "MX",
		Data: "7 127.0.0.1",
		TTL:  time.Duration(3600 * time.Second),
	}

	ikRec := ToInfomaniakRecord(&libRec, &ZoneMapping{InfomaniakManagedZone: "domain.com", LibDnsZone: "test.domain.com"})
	assertEquals(t, "Source", "test", ikRec.Source)
	assertEquals(t, "Type", "MX", ikRec.Type)
	assertEquals(t, "Target", "7 127.0.0.1", ikRec.Target)
	assertEqualsInt(t, "TTL", 3600, ikRec.TtlInSec)
}

func Test_ToInfomaniakRecord_DefaultTtlIsAppliedIfNoTtlProvided(t *testing.T) {
	ikRec := ToInfomaniakRecord(&libdns.RR{}, &ZoneMapping{})
	assertEqualsInt(t, "TTL", 300, ikRec.TtlInSec)
}
