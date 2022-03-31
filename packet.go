package uotp

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

type packet struct {
	status     status
	payload    payload
	oid        uint64
	sharedKey  []byte
	extraToken []byte
}

func newPacket(opcode opCode) *packet {
	var payload payload
	switch opcode {
	case opCodeTime:
		payload = new(payloadTime)
	case opCodeIssue:
		payload = new(payloadIssue)
	case opCodeResetErrorCount:
		payload = new(payloadResetErrorCount)
	case opCodeInformation:
		payload = new(payloadInfomation)
	case opCodeUseHistory:
		payload = new(History)
	case opCodeHelp:
		payload = new(payloadHelp)
	}

	p := &packet{
		status:  statusOK,
		payload: payload,
	}
	payload.initPacket(p)

	return p
}

func (p *packet) Send(ctx context.Context) (*packet, error) {
	cryptoKey := p.getCryptoKey()

	buf, err := encodePacket(p, cryptoKey)
	if err != nil {
		return nil, err
	}

	// Cancel net.Conn
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", "211.49.97.230:20004")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(buf)
	if err != nil {
		return nil, err
	}

	if len(buf) < 6 {
		buf = make([]byte, 6)
	}
	_, err = io.ReadFull(conn, buf[:6])
	if err != nil && err != io.EOF {
		return nil, err
	}

	dataSize, err := strconv.Atoi(b2s(buf[1:6]))
	if err != nil {
		return nil, err
	}
	if len(buf) < dataSize {
		buf = make([]byte, dataSize)
	}

	_, err = io.ReadFull(conn, buf[:dataSize])
	if err != nil && err != io.EOF {
		return nil, err
	}

	err = conn.Close()
	if err != nil {
		return nil, err
	}

	return decodePacket(buf[:dataSize], cryptoKey)
}

func (p *packet) setEncryptionInfo(sharedKey []byte, extraToken string) {
	if sharedKey != nil {
		p.sharedKey = append(p.sharedKey[:0], sharedKey...)
	}
	if extraToken != "" {
		p.extraToken = s2b(fmt.Sprintf("%07s ", extraToken))
	}
}

func (p *packet) getCryptoKey() []byte {
	if len(p.sharedKey) > 0 {
		if len(p.extraToken) > 0 {
			sharedKeyDecoded := make([]byte, len(p.sharedKey)/2)
			hex.Decode(sharedKeyDecoded, p.sharedKey)

			h := sha1.New()
			h.Write(sharedKeyDecoded)
			h.Write(p.extraToken)
			return h.Sum(nil)
		} else {
			r := make([]byte, len(p.sharedKey))
			copy(r, p.sharedKey)
			return r
		}
	} else {
		return nil
	}
}

func (p *packet) appendCommonHeader(w io.Writer) {
	r := rand.Intn(3)

	switch r {
	case 0:
		fmt.Fprintf(w, "%-3s", "KTF")
	case 1:
		fmt.Fprintf(w, "%-3s", "SKT")
	case 2:
		fmt.Fprintf(w, "%-3s", "LGT")
	}
	if p.oid != 0 {
		fmt.Fprintf(w, "%-11d", p.oid)
	} else {
		fmt.Fprintf(w, "%-11s", "")
	}
	switch r {
	case 0:
		fmt.Fprintf(w, "%-16s", "SM-G920K")
	case 1:
		fmt.Fprintf(w, "%-16s", "SM-G950S")
	case 2:
		fmt.Fprintf(w, "%-16s", "SM-G955L")
	}
	fmt.Fprintf(w, "%-4s", "GA15")
	fmt.Fprintf(w, "%04d", 2)
	fmt.Fprintf(w, "%04d", 0)
}

func encodePacket(p *packet, cryptoKey []byte) ([]byte, error) {
	var payload bytes.Buffer
	if p.payload.needsCommonHeader() {
		p.appendCommonHeader(&payload)
	}
	p.payload.encode(&payload)
	payload.Write(p.extraToken)

	payloadData := payload.Bytes()

	if len(cryptoKey) != 0 && payloadData != nil {
		var err error
		payloadData, err = encrypt(cryptoKey, payloadData)
		if err != nil {
			return nil, err
		}
	}

	var sharedKey [64]byte
	for i := 0; i < 64; i++ {
		sharedKey[i] = ' '
	}
	if len(p.sharedKey) > 0 {
		copy(sharedKey[64-len(p.sharedKey):], p.sharedKey)
	}

	bodyLen := len(sharedKey) + 4 + 3 + len(payloadData)

	var data bytes.Buffer
	data.Grow(1 + 5 + bodyLen)
	fmt.Fprintf(&data, "S%05d", bodyLen)
	data.Write(sharedKey[:])
	fmt.Fprintf(&data, "%04s%03d", string(p.status), int(p.payload.opcode()))
	data.Write(payloadData)

	return data.Bytes(), nil
}

func decodePacket(data []byte, cryptoKey []byte) (pnew *packet, err error) {
	if len(data) < 71 {
		return nil, ErrInvalidPacket
	}

	sharedKeyRaw := data[:64]
	statusStrRaw := data[64:68]
	opcodeRaw := data[68:71]

	sharedKey, err := hex.DecodeString(strings.TrimRight(b2s(sharedKeyRaw), " "))
	if err != nil && err != io.EOF {
		return nil, ErrInvalidPacket
	}

	opcodeInt, err := strconv.Atoi(string(opcodeRaw))
	if err != nil {
		return nil, ErrInvalidPacket
	}

	pnew = newPacket(opCode(opcodeInt))
	pnew.status = status(string(statusStrRaw))
	pnew.sharedKey = sharedKey

	payload := data[71:]

	if len(pnew.sharedKey) != 0 || len(cryptoKey) != 0 {
		if len(pnew.sharedKey) != 0 {
			payload, err = decrypt(pnew.sharedKey, payload)
		} else {
			payload, err = decrypt(cryptoKey, payload)
		}
		if err != nil {
			return nil, ErrInvalidPacket
		}
	}

	if pnew.status != statusOK {
		return nil, fmt.Errorf("status: %s, payload: %s", pnew.status, decodeEUCKR(payload))
	}

	err = pnew.payload.decode(payload)
	if err != nil {
		return nil, err
	}

	return pnew, nil
}

func newSharedKey() []byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], otpNow())

	h := sha256.New()
	h.Write(b[:])
	var sb bytes.Buffer
	hex.NewEncoder(&sb).Write(h.Sum(nil))

	return sb.Bytes()
}
