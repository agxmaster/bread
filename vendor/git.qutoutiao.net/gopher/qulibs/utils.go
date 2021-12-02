package qulibs

import (
	"os"
	"unsafe"
)

const (
	CIEnvKey = "PEDESTAL_CI"
)

func IsCI() bool {
	switch os.Getenv(CIEnvKey) {
	case "yes", "on":
		return true
	}

	return false
}

// String converts byte slice to string.
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Bytes converts string to byte slice.
func Bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
