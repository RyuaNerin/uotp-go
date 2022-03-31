package uotp

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type History struct {
	requestPage   int
	requestPeriod int

	PeriodStart time.Time
	PeriodEnd   time.Time

	PageCurrent int
	PageTotal   int
	Entries     []HistoryEntry
}

type HistoryEntry struct {
	At   time.Time
	Type string
	Name string
}

func (p *History) opcode() opCode {
	return opCodeUseHistory
}
func (p *History) needsCommonHeader() bool {
	return true
}
func (p *History) initPacket(pk *packet) {
}
func (p *History) encode(w io.Writer) {
	fmt.Fprintf(w, "%04d%1d", p.requestPage, p.requestPeriod)
}
func (p *History) decode(payload []byte) (err error) {
	p.PeriodStart, err = time.ParseInLocation("2006-01-02", string(payload[0:10]), time.Local)
	if err != nil {
		return ErrInvalidPacket
	}

	p.PeriodEnd, err = time.ParseInLocation("2006-01-02", string(payload[10:10+10]), time.Local)
	if err != nil {
		return ErrInvalidPacket
	}

	p.PageTotal, err = strconv.Atoi(string(payload[20 : 20+4]))
	if err != nil {
		return ErrInvalidPacket
	}

	p.PageCurrent, err = strconv.Atoi(string(payload[20+4 : 20+4+4]))
	if err != nil {
		return ErrInvalidPacket
	}

	dataCount, err := strconv.Atoi(string(payload[20+4+4 : 20+4+4+2]))
	if err != nil {
		return ErrInvalidPacket
	}
	if len(payload) < 30+(18+40+40)*dataCount {
		return ErrInvalidPacket
	}

	offset := 30
	p.Entries = make([]HistoryEntry, 0, dataCount)
	for i := 0; i < int(dataCount); i++ {
		date, err := time.ParseInLocation("2006-01-0215:04:05", b2s(payload[offset:offset+18]), time.Local)
		if err != nil {
			return ErrInvalidPacket
		}

		type_ := strings.TrimSpace(decodeEUCKR(payload[offset+18 : offset+18+40]))
		name := strings.TrimSpace(decodeEUCKR(payload[offset+18+40 : offset+18+40+40]))

		p.Entries = append(
			p.Entries,
			HistoryEntry{
				At:   date,
				Type: type_,
				Name: name,
			},
		)
		offset += 18 + 40 + 40
	}

	return nil
}
