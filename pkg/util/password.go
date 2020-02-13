package util

import (
	"github.com/alexandrevicenzi/unchained"
	"github.com/alexandrevicenzi/unchained/pbkdf2"
)

var (
	hasher = pbkdf2.NewPBKDF2SHA256Hasher()
)

// MakePassword creates pbkdf2 sha256 based password string.
func MakePassword(password string) string {
	encoded, _ := hasher.Encode(password, GenerateRandomString(12), 100000)
	return encoded
}

// VerifyPassword verifies pbkdf2 sha256 encoded password.
func VerifyPassword(password, encoded string) bool {
	if !unchained.IsPasswordUsable(encoded) {
		return false
	}

	ok, _ := hasher.Verify(password, encoded)

	return ok
}
