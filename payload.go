package uotp

import (
	"errors"
	"io"
)

type status string

const (
	statusOK status = "0000"
)

type opCode int

const (
	opCodeError           opCode = -1
	opCodeInformation     opCode = 402
	opCodeTime            opCode = 407
	opCodeIssue           opCode = 451
	opCodeResetErrorCount opCode = 452
	opCodeUseHistory      opCode = 453
	opCodeHelp            opCode = 454
)

type payload interface {
	opcode() opCode
	needsCommonHeader() bool
	initPacket(p *packet)
	encode(w io.Writer)
	decode(payload []byte) error
}

var ErrInvalidPacket = errors.New("invalid packet")
