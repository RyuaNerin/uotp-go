package uotp

import (
	"encoding/binary"
	"io"
)

type payloadTime struct {
	Time int
}

func (p *payloadTime) opcode() opCode {
	return opCodeTime
}
func (p *payloadTime) needsCommonHeader() bool {
	return false
}
func (p *payloadTime) initPacket(pk *packet) {
}
func (p *payloadTime) encode(w io.Writer) {
}
func (p *payloadTime) decode(payload []byte) error {
	if len(payload) < 4 {
		return ErrInvalidPacket
	}
	p.Time = int(binary.BigEndian.Uint32(payload))
	return nil
}
