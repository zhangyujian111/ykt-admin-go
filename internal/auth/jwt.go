package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明（替换 sa-token 的 tokenInfo）。
type Claims struct {
	UserID   int64    `json:"uid"`
	Username string   `json:"username"`
	TenantID int64    `json:"tid"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func New(secret, issuer string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

// Issue 签发 token。
func (m *Manager) Issue(userID int64, username string, tenantID int64, roles []string) (string, error) {
	now := time.Now()
	c := Claims{
		UserID:   userID,
		Username: username,
		TenantID: tenantID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString(m.secret)
}

// Parse 解析 token。
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	c := &Claims{}
	t, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("token invalid")
	}
	return c, nil
}
