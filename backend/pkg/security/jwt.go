package security

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"employee-portal/modules/entity"
)

type claims struct {
	EmployeeID         int64       `json:"emp,omitempty"`
	Role               entity.Role `json:"role"`
	MustChangePassword bool        `json:"mcp,omitempty"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (m *TokenManager) TTL() time.Duration { return m.ttl }

func (m *TokenManager) Issue(a entity.Actor) (string, time.Time, error) {
	now := m.now()
	exp := now.Add(m.ttl)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		EmployeeID:         a.EmployeeID,
		Role:               a.Role,
		MustChangePassword: a.MustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(a.UserID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	})
	s, err := tok.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return s, exp, nil
}

func (m *TokenManager) Parse(token string) (entity.Actor, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return entity.Actor{}, entity.ErrUnauthorized
	}
	uid, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil || uid <= 0 {
		return entity.Actor{}, entity.ErrUnauthorized
	}
	if c.Role != entity.RoleEmployee && c.Role != entity.RoleAdmin {
		return entity.Actor{}, errors.Join(entity.ErrUnauthorized, errors.New("unknown role"))
	}
	return entity.Actor{UserID: uid, EmployeeID: c.EmployeeID, Role: c.Role, MustChangePassword: c.MustChangePassword}, nil
}
