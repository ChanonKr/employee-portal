package security

import (
	"crypto/rand"
	"math/big"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

// dummyHash is compared when the user does not exist, so a login attempt takes
// the same time whether or not the employee code is valid.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-equalizer"), bcryptCost)

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(h), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// BurnPasswordCheck spends the same CPU as CheckPassword when there is no user.
func BurnPasswordCheck(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
}

// Characters for temporary passwords: no 0/O, 1/l/I, so they are easy to read out.
const tempAlphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// TempPassword returns a random 12-character password with at least one letter
// and one digit, so it passes the password policy.
func TempPassword() (string, error) {
	const n = 12
	for {
		b := make([]byte, n)
		for i := range b {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(tempAlphabet))))
			if err != nil {
				return "", err
			}
			b[i] = tempAlphabet[idx.Int64()]
		}
		s := string(b)
		if strings.ContainsAny(s, "23456789") && strings.IndexFunc(s, unicode.IsLetter) >= 0 {
			return s, nil
		}
	}
}
