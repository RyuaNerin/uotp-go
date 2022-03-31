package uotp

import (
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

func decodeEUCKR(data []byte) string {
	s, _, _ := transform.String(korean.EUCKR.NewDecoder(), string(data))
	return s
}
