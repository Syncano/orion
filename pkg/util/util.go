package util

import (
	"context"
	"errors"
	"hash/fnv"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const hexBytes = "abcdef0123456789"

// IsTrue returns true if string is a true value.
func IsTrue(s string) bool {
	s = strings.ToLower(s)
	return s == "1" || s == "true" || s == "t" || s == "yes"
}

func generateRandomString(n int, base string) string {
	b := make([]byte, n)
	l := len(base)

	for i := range b {
		b[i] = base[rand.Intn(l)]
	}

	return string(b)
}

// GenerateRandomString creates random string with a-zA-Z0-9 of specified length.
func GenerateRandomString(n int) string {
	return generateRandomString(n, letterBytes)
}

// GenerateRandomHexString creates random string with a-f0-9 of specified length.
func GenerateRandomHexString(n int) string {
	return generateRandomString(n, hexBytes)
}

// GenerateKey creates random string with length=40.
func GenerateKey() string {
	return GenerateRandomString(40)
}

// GenerateHexKey creates random hex string with length=40.
func GenerateHexKey() string {
	return GenerateRandomHexString(40)
}

// GenerateHexKeyWithParity creates random hex string with defined parity and length=40.
func GenerateHexKeyWithParity(parity bool) string {
	b := []rune(GenerateHexKey())
	if parity {
		b[len(b)-1] = b[len(b)-1] & 14
	} else {
		b[len(b)-1] = b[len(b)-1] | 1
	}

	return string(b)
}

// CheckStringParity checks if passed string key is odd or even (considering last bit).
func CheckStringParity(s string) bool {
	b := []rune(s)
	i, _ := strconv.ParseInt(string(b[len(b)-1]), 8, 32) // nolint: errcheck

	return i&1 != 1
}

// Retry retries f() and sleeps between retries.
func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)
	}

	return
}

// RetryWithCritical is similar to Retry but allows func to return true/false to mark if returned error is critical or not.
func RetryWithCritical(attempts int, sleep time.Duration, f func() (bool, error)) (critical bool, err error) {
	for i := 0; ; i++ {
		critical, err = f()
		if err == nil || critical {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)
	}

	return
}

func RetryNotCancelled(attempts int, sleep time.Duration, f func() error) (bool, error) {
	canceled, err := RetryWithCritical(attempts, sleep, func() (bool, error) {
		err := f()
		return IsContextError(err), err
	})

	return canceled, err
}

func checkError(err, target error, code codes.Code) bool {
	if errors.Is(err, target) {
		return true
	}

	s, ok := status.FromError(err)

	return ok && s.Code() == code
}

func IsContextError(err error) bool {
	return IsCancellation(err) || IsDeadlineExceeded(err)
}

func IsCancellation(err error) bool {
	return checkError(err, context.Canceled, codes.Canceled)
}

func IsDeadlineExceeded(err error) bool {
	return checkError(err, context.DeadlineExceeded, codes.DeadlineExceeded)
}

// Must is a small helper that checks if err is nil. If it's not - panics.
// This should be used only for errors that can never actually happen and are impossible to simulate in tests.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Hash returns FNV 1a hash of string.
func Hash(s string) uint32 {
	h := fnv.New32a()
	_, e := h.Write([]byte(s))

	Must(e)

	return h.Sum32()
}

const lowerhex = "0123456789abcdef"

// ToQuoteJSON converts source []byte to ASCII only quoted []byte.
func ToQuoteJSON(s []byte) []byte { // nolint: gocyclo
	buf := make([]byte, 0, 3*len(s)/2)
	buf = append(buf, '"')

	var width int

	for ; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1

		if r == '"' || r == '\\' {
			buf = append(buf, '\\', byte(r))
			continue
		}

		if r < utf8.RuneSelf && strconv.IsPrint(r) {
			buf = append(buf, byte(r))
			continue
		}

		r, width = utf8.DecodeRune(s)
		switch r {
		case '\b':
			buf = append(buf, `\b`...)
		case '\f':
			buf = append(buf, `\f`...)
		case '\n':
			buf = append(buf, `\n`...)
		case '\r':
			buf = append(buf, `\r`...)
		case '\t':
			buf = append(buf, `\t`...)
		default:
			switch {
			case r < 0x10000:
				buf = append(buf, `\u`...)

				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			default:
				r -= 0x10000
				r = (r >> 10) + 0xd800

				buf = append(buf, `\u`...)

				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}

				r = (r & 0x3ff) + 0xdc00

				buf = append(buf, `\u`...)

				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			}
		}
	}

	buf = append(buf, '"')

	return buf
}

// NonEmptyString returns first non empty string.
func NonEmptyString(a, b string) string {
	if a == "" {
		return b
	}

	return a
}

// Truncate returns string truncated to length i.
func Truncate(a string, i int) string {
	if len(a) > i {
		return a[:i]
	}

	return a
}

// RegexNamedGroups returns regex groups for specified regex and matches list.
func RegexNamedGroups(regex *regexp.Regexp, matches []string) map[string]string {
	m, n := matches[1:], regex.SubexpNames()[1:]
	r := make(map[string]string, len(m))

	for i := range n {
		if n[i] != "" {
			r[n[i]] = m[i]
		}
	}

	return r
}
