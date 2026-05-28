package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"atom-maintenance/internal/adapters/http/dto"
	"atom-maintenance/internal/app"
	"atom-maintenance/internal/domain"
	"atom-maintenance/pkg/authctx"
	"atom-maintenance/pkg/respond"
	"atom-maintenance/platform/logger"

	"github.com/go-playground/validator/v10"
)

type AuthHandlers struct {
	auth     *app.AuthApp
	log      *slog.Logger
	validate *validator.Validate
}

func NewAuthHandlers(auth *app.AuthApp, log *slog.Logger) *AuthHandlers {
	return &AuthHandlers{auth: auth, log: log, validate: validator.New()}
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.auth", "op", "Login")
	start := time.Now()

	var in dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	pair, err := h.auth.Login(ctx, in.Email, in.Password)
	if err != nil {
		log.Warn("login failed",
			"email", in.Email,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("login success",
		"email", in.Email,
		"duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, dto.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.auth", "op", "Refresh")
	start := time.Now()

	var in dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Warn("decode body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}
	if err := h.validate.Struct(in); err != nil {
		log.Warn("validate body", "err", err)
		respond.Error(w, domain.ErrBadRequest)
		return
	}

	pair, err := h.auth.Refresh(ctx, in.RefreshToken)
	if err != nil {
		log.Warn("refresh failed",
			"err", err,
			"duration_ms", time.Since(start).Milliseconds())
		respond.Error(w, err)
		return
	}

	log.Info("tokens refreshed", "duration_ms", time.Since(start).Milliseconds())
	respond.JSON(w, http.StatusOK, dto.RefreshResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.WithReqID(ctx, h.log).With("module", "http.auth", "op", "Logout")
	start := time.Now()

	p, ok := authctx.PrincipalFrom(ctx)
	if !ok {
		respond.Error(w, domain.ErrUnauthorized)
		return
	}

	_ = h.auth.Logout(ctx, p.EmployeeID, p.JTI, p.AccessTTL)

	log.Info("logout success",
		"employee_id", p.EmployeeID,
		"duration_ms", time.Since(start).Milliseconds())
	w.WriteHeader(http.StatusNoContent)
}
