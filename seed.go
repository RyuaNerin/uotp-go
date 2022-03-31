package uotp

import (
	"crypto/cipher"
	"crypto/sha1"

	"github.com/RyuaNerin/uotp/seed"
)

func populateKeyAndIV(key []byte, iv []byte, sharedKey []byte) {
	_ = key[15]
	_ = iv[15]

	h := sha1.New()

	h.Write(sharedKey)
	r := h.Sum(nil)

	for i := 0; i < 4; i++ {
		h.Reset()
		h.Write(r)
		r = h.Sum(r[:0])
	}

	h.Reset()
	h.Write(r[16:])
	copy(iv, h.Sum(nil)[:16])
	copy(key, r[:16])
}

func encrypt(sharedKey []byte, src []byte) ([]byte, error) {
	var key, iv [16]byte
	populateKeyAndIV(key[:], iv[:], sharedKey)

	b, err := seed.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	padSize := 16 - (len(src) % 16)
	srcPadded := make([]byte, len(src)+padSize)
	copy(srcPadded, src)
	for i := len(srcPadded) - padSize; i < len(srcPadded); i++ {
		srcPadded[i] = byte(padSize)
	}

	c := cipher.NewCBCEncrypter(b, iv[:])

	dst := make([]byte, len(srcPadded))
	c.CryptBlocks(dst, srcPadded)

	return dst, nil
}

func decrypt(sharedKey []byte, src []byte) ([]byte, error) {
	var key, iv [16]byte
	populateKeyAndIV(key[:], iv[:], sharedKey)

	b, err := seed.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	c := cipher.NewCBCDecrypter(b, iv[:])

	dst := make([]byte, len(src))
	c.CryptBlocks(dst, src)

	pad := int(dst[len(dst)-1])
	if pad < 16 {
		isPadded := true
		for i := len(dst) - pad; i < len(dst); i++ {
			if dst[i] != byte(pad) {
				isPadded = false
				break
			}
		}

		if isPadded {
			dst = dst[:len(dst)-pad]
		}
	}

	return dst, nil
}
