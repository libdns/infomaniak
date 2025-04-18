package infomaniak

import (
	"strconv"
	"time"

	"fmt"

	"github.com/libdns/libdns"
)

// Default priority - infomaniak does not return any value
const defaultPriority = 10

// Default TTL that is applied if none is provided - infomaniak requires a TTL
const defaultTtlSecs = 300

// ToLibDnsRecord maps a infomaniak dns record to a libdns record
func (ikr *IkRecord) ToLibDnsRecord() libdns.Record {
	return libdns.Record{
		ID:       fmt.Sprint(ikr.ID),
		Type:     ikr.Type,
		Name:     ikr.Source,
		Value:    ikr.Target,
		TTL:      time.Duration(ikr.TtlInSec),
		Priority: defaultPriority,
	}
}

// ToInfomaniakRecord maps a libdns record to a infomaniak dns record
func ToInfomaniakRecord(libdnsRec *libdns.Record) IkRecord {
	ikRec := IkRecord{
		ID:       0,
		Type:     libdnsRec.Type,
		Source:   libdnsRec.Name,
		Target:   libdnsRec.Value,
		TtlInSec: int(libdnsRec.TTL),
	}

	id, err := strconv.Atoi(libdnsRec.ID)
	if err == nil {
		ikRec.ID = id
	}

	if ikRec.TtlInSec <= 0 {
		ikRec.TtlInSec = defaultTtlSecs
	}
	return ikRec
}
