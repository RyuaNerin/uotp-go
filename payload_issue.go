package uotp

import (
	"encoding/hex"
	"io"
	"strconv"
	"strings"
)

type payloadIssue struct {
	oid          uint64
	seed         []byte
	serialNumber string
	userHash     string
	issueInfo    string
}

func (p *payloadIssue) opcode() opCode {
	return opCodeIssue
}
func (p *payloadIssue) needsCommonHeader() bool {
	return true
}
func (p *payloadIssue) initPacket(pk *packet) {
	pk.setEncryptionInfo(newSharedKey(), "")
}
func (p *payloadIssue) encode(w io.Writer) {
}
func (p *payloadIssue) decode(payload []byte) error {
	if len(payload) < 20+11+40+64+80 {
		return ErrInvalidPacket
	}

	oid, err := strconv.ParseUint(string(payload[20:20+11]), 10, 64)
	if err != nil {
		return ErrInvalidPacket
	}

	seed := make([]byte, 20)
	_, err = hex.Decode(seed, payload[20+11:20+11+40])
	if err != nil {
		return ErrInvalidPacket
	}

	p.serialNumber = strings.TrimSpace(string(payload[0:20]))
	p.oid = uint64(oid)
	p.seed = seed
	p.userHash = string(payload[20+11+40 : 20+11+40+64])
	p.issueInfo = strings.TrimSpace(string(payload[20+11+40+64 : 20+11+40+64+80]))

	return nil
}
