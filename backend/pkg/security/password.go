package security

import "golang.org/x/crypto/bcrypt"

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
