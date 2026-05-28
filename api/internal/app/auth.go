package app

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"atom-maintenance/internal/domain"
	"atom-maintenance/internal/ports"
	jwtpkg "atom-maintenance/pkg/jwt"
	"atom-maintenance/pkg"
	"atom-maintenance/platform/logger"
)

const (
	maxFailedAttempts = 5
	refreshKeyPrefix  = "refresh:"
	blacklistPrefix   = "blacklist:"
)

type AuthApp struct {
	employees domain.EmployeeRepository
	cache     ports.Cache
	jwt       *jwtpkg.Provider
	log       *slog.Logger
}

func NewAuthApp(employees domain.EmployeeRepository, cache ports.Cache, jwt *jwtpkg.Provider, log *slog.Logger) *AuthApp {
	return &AuthApp{employees: employees, cache: cache, jwt: jwt, log: log}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (a *AuthApp) Login(ctx context.Context, email, password string) (TokenPair, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.auth", "op", "Login")

	e, err := a.employees.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Warn("login attempt: email not found", "email", email)
			return TokenPair{}, domain.ErrInvalidCredentials
		}
		return TokenPair{}, err
	}

	if e.DeletedAt != nil {
		log.Warn("login attempt for deleted/blocked employee", "id", e.ID, "email", email)
		return TokenPair{}, domain.ErrForbidden
	}

	if !pkg.CheckPassword(e.PasswordHash, password) {
		_ = a.employees.IncrFailedAttempts(ctx, e.ID)
		attempts := e.FailedAttempts + 1
		if attempts >= maxFailedAttempts {
			_ = a.employees.SoftDelete(ctx, e.ID)
			log.Warn("employee auto-blocked after failed attempts", "id", e.ID, "attempts", attempts)
		} else {
			log.Warn("login attempt: wrong password", "id", e.ID, "email", email, "failed_attempts", attempts)
		}
		return TokenPair{}, domain.ErrInvalidCredentials
	}

	_ = a.employees.ResetFailedAttempts(ctx, e.ID)

	accessToken, _, _, err := a.jwt.GenerateAccess(e.ID, e.Role)
	if err != nil {
		log.Error("generate access token", "err", err)
		return TokenPair{}, domain.ErrInternal
	}

	refreshToken, _, err := a.jwt.GenerateRefresh(e.ID)
	if err != nil {
		log.Error("generate refresh token", "err", err)
		return TokenPair{}, domain.ErrInternal
	}

	if err := a.cache.Set(ctx, refreshKey(e.ID), refreshToken, 720*time.Hour); err != nil {
		log.Error("cache set refresh token", "err", err)
		return TokenPair{}, domain.ErrInternal
	}

	log.Info("employee logged in", "id", e.ID)
	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (a *AuthApp) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "app.auth", "op", "Refresh")

	employeeID, _, err := a.jwt.VerifyRefresh(refreshToken)
	if err != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}

	stored, err := a.cache.Get(ctx, refreshKey(employeeID))
	if err != nil {
		if errors.Is(err, domain.ErrCacheMiss) {
			return TokenPair{}, domain.ErrUnauthorized
		}
		return TokenPair{}, domain.ErrInternal
	}

	if subtle.ConstantTimeCompare([]byte(stored), []byte(refreshToken)) != 1 {
		return TokenPair{}, domain.ErrUnauthorized
	}

	e, err := a.employees.GetByID(ctx, employeeID)
	if err != nil {
		return TokenPair{}, domain.ErrUnauthorized
	}
	if e.DeletedAt != nil {
		return TokenPair{}, domain.ErrForbidden
	}

	newAccess, _, _, err := a.jwt.GenerateAccess(e.ID, e.Role)
	if err != nil {
		return TokenPair{}, domain.ErrInternal
	}

	newRefresh, _, err := a.jwt.GenerateRefresh(e.ID)
	if err != nil {
		return TokenPair{}, domain.ErrInternal
	}

	if err := a.cache.Set(ctx, refreshKey(e.ID), newRefresh, 720*time.Hour); err != nil {
		return TokenPair{}, domain.ErrInternal
	}

	log.Info("tokens refreshed", "employee_id", e.ID)
	return TokenPair{AccessToken: newAccess, RefreshToken: newRefresh}, nil
}

func (a *AuthApp) Logout(ctx context.Context, employeeID int64, jti string, accessTTL time.Duration) error {
	log := logger.WithReqID(ctx, a.log).With("module", "app.auth", "op", "Logout")

	_ = a.cache.Delete(ctx, refreshKey(employeeID))

	if accessTTL > 0 {
		if err := a.cache.Set(ctx, blacklistKey(jti), "revoked", accessTTL); err != nil {
			log.Error("blacklist access token", "err", err)
		}
	}

	log.Info("employee logged out", "employee_id", employeeID)
	return nil
}

func refreshKey(employeeID int64) string { return fmt.Sprintf("%s%d", refreshKeyPrefix, employeeID) }
func blacklistKey(jti string) string     { return blacklistPrefix + jti }
