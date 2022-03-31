package uotp

import (
	"io"
	"strconv"
	"strings"
)

type payloadInfomation struct {
	oid     int
	seeed   string
	partner string
}

func (p *payloadInfomation) opcode() opCode {
	return opCodeInformation
}
func (p *payloadInfomation) needsCommonHeader() bool {
	return true
}
func (p *payloadInfomation) initPacket(pk *packet) {
}
func (p *payloadInfomation) encode(w io.Writer) {
}
func (p *payloadInfomation) decode(payload []byte) error {
	if len(payload) < 11+40+80 {
		return ErrInvalidPacket
	}

	oid, err := strconv.ParseInt(string(payload[:11]), 10, 32)
	if err != nil {
		return ErrInvalidPacket
	}

	p.oid = int(oid)
	p.seeed = strings.TrimSpace(string(payload[11 : 11+40]))
	p.partner = strings.TrimSpace(string(payload[11+40 : 11+40+80]))

	return nil
}
