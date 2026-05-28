package jwt

import (
	"errors"
	"fmt"
	"time"

	"atom-maintenance/internal/config"
	"atom-maintenance/internal/domain"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type claims struct {
	gojwt.RegisteredClaims
	Type string `json:"typ"`
	Role string `json:"role,omitempty"`
}

type Provider struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(cfg config.JWTConfig) *Provider {
	return &Provider{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}
}

func (p *Provider) GenerateAccess(employeeID int64, role domain.EmployeeRole) (token, jti string, ttl time.Duration, err error) {
	jti = uuid.NewString()
	now := time.Now()
	exp := now.Add(p.accessTTL)

	c := claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", employeeID),
			ID:        jti,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(exp),
		},
		Type: tokenTypeAccess,
		Role: string(role),
	}

	token, err = gojwt.NewWithClaims(gojwt.SigningMethodHS256, c).SignedString(p.secret)
	return token, jti, p.accessTTL, err
}

func (p *Provider) GenerateRefresh(employeeID int64) (token, jti string, err error) {
	jti = uuid.NewString()
	now := time.Now()

	c := claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", employeeID),
			ID:        jti,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(p.refreshTTL)),
		},
		Type: tokenTypeRefresh,
	}

	token, err = gojwt.NewWithClaims(gojwt.SigningMethodHS256, c).SignedString(p.secret)
	return token, jti, err
}

func (p *Provider) VerifyAccess(tokenStr string) (employeeID int64, role domain.EmployeeRole, jti string, ttl time.Duration, err error) {
	c, err := p.parse(tokenStr)
	if err != nil {
		return 0, "", "", 0, domain.ErrUnauthorized
	}
	if c.Type != tokenTypeAccess {
		return 0, "", "", 0, domain.ErrUnauthorized
	}
	if c.Role == "" {
		return 0, "", "", 0, domain.ErrUnauthorized
	}

	id, err := parseSubject(c)
	if err != nil {
		return 0, "", "", 0, domain.ErrUnauthorized
	}

	remaining := time.Until(c.ExpiresAt.Time)
	return id, domain.EmployeeRole(c.Role), c.ID, remaining, nil
}

func (p *Provider) VerifyRefresh(tokenStr string) (employeeID int64, jti string, err error) {
	c, err := p.parse(tokenStr)
	if err != nil {
		return 0, "", domain.ErrUnauthorized
	}
	if c.Type != tokenTypeRefresh {
		return 0, "", domain.ErrUnauthorized
	}

	id, err := parseSubject(c)
	if err != nil {
		return 0, "", domain.ErrUnauthorized
	}

	return id, c.ID, nil
}

func (p *Provider) parse(tokenStr string) (*claims, error) {
	t, err := gojwt.ParseWithClaims(tokenStr, &claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return p.secret, nil
	})
	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, domain.ErrUnauthorized
		}
		return nil, domain.ErrUnauthorized
	}

	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return nil, domain.ErrUnauthorized
	}
	return c, nil
}

func parseSubject(c *claims) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(c.Subject, "%d", &id)
	return id, err
}
