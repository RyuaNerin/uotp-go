package uotp

import (
	"io"
	"strings"
)

type payloadHelp struct {
	messages []string
}

func (p *payloadHelp) opcode() opCode {
	return opCodeHelp
}
func (p *payloadHelp) needsCommonHeader() bool {
	return true
}
func (p *payloadHelp) initPacket(pk *packet) {
}
func (p *payloadHelp) encode(w io.Writer) {
}
func (p *payloadHelp) decode(payload []byte) error {
	if len(payload) < 8 {
		return ErrInvalidPacket
	}

	payload = payload[:len(payload)-8]
	p.messages = strings.Split(decodeEUCKR(payload), "|")

	return nil
}
