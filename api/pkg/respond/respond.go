package respond

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"atom-maintenance/internal/domain"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("respond: encode json", "err", err)
	}
}

func Error(w http.ResponseWriter, err error) {
	code, status, msg := mapError(err)
	JSON(w, status, ErrorResponse{Code: code, Message: msg})
}

func mapError(err error) (code string, status int, msg string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "NotFound", http.StatusNotFound, "Ресурс не найден"
	case errors.Is(err, domain.ErrConflict):
		return "Conflict", http.StatusConflict, "Ресурс уже существует или нарушена уникальность данных"
	case errors.Is(err, domain.ErrBadRequest):
		return "BadRequest", http.StatusBadRequest, "Неверный формат или значения данных запроса"
	case errors.Is(err, domain.ErrUnauthorized):
		return "Unauthorized", http.StatusUnauthorized, "Сессия неактивна или неверна"
	case errors.Is(err, domain.ErrForbidden):
		return "Forbidden", http.StatusForbidden, "Недостаточно прав для выполнения операции"
	case errors.Is(err, domain.ErrInvalidCredentials):
		return "InvalidCredentials", http.StatusUnauthorized, "Неверный email или пароль"
	case errors.Is(err, domain.ErrTooManyRequests):
		return "TooManyRequests", http.StatusTooManyRequests, "Слишком много запросов, попробуйте позже"
	case errors.Is(err, domain.ErrChecklistIncomplete):
		return "ChecklistIncomplete", http.StatusBadRequest, err.Error()
	default:
		return "InternalError", http.StatusInternalServerError, "Внутренняя ошибка сервера, попробуйте позже"
	}
}
