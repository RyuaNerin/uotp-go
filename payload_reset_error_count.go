package uotp

import "io"

type payloadResetErrorCount struct {
}

func (p *payloadResetErrorCount) opcode() opCode {
	return opCodeResetErrorCount
}
func (p *payloadResetErrorCount) needsCommonHeader() bool {
	return true
}
func (p *payloadResetErrorCount) initPacket(pk *packet) {
}
func (p *payloadResetErrorCount) encode(w io.Writer) {
}
func (p *payloadResetErrorCount) decode(payload []byte) error {
	return nil
}
