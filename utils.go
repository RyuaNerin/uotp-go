package uotp

import (
	"math"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"
)

func otpNow() uint32 {
	now := time.Now()

	return uint32(
		(now.Year()-2000)*31536000 +
			(int(now.Month())-1)*2592000 +
			(int(now.Day())-1)*86400 +
			now.Hour()*3600 +
			now.Minute()*60 +
			now.Second(),
	)
}

// maxgroup = -1
func humanize(text string, char string, each int, maxgroup int) string {
	if maxgroup == -1 {
		maxgroup = int(math.Ceil(float64(utf8.RuneCountInString(text)) / float64(each)))
	}

	var sb strings.Builder

	charCount := 0
	groupCount := 1
	for len(text) > 0 {
		r, sz := utf8.DecodeRuneInString(text)
		text = text[sz:]
		sb.WriteRune(r)
		charCount++

		if groupCount < maxgroup && charCount == each {
			charCount = 0
			sb.WriteString(char)

			groupCount++
		}
	}

	return sb.String()
}

func b2s(b []byte) (s string) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sh.Data = bh.Data
	sh.Len = bh.Len
	return s
}

func s2b(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}
