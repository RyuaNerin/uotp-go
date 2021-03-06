package seed

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestEncrypt(t *testing.T) {
	key := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	data := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
	}

	b, _ := NewCipher(key)

	dstEnc := make([]byte, 16)
	dstDec := make([]byte, 16)
	b.Encrypt(dstEnc, data)
	b.Decrypt(dstDec, dstEnc)

	if !bytes.Equal(data, dstDec) {
		t.Errorf("Data : %s\nEnc  : %s\nDec  : %s", hex.Dump(data), hex.Dump(dstEnc), hex.Dump(dstDec))
	}
}
