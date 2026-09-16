package transport

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type WebResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WriteJSON writes a standard JSON response to the ResponseWriter.
func WriteJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := WebResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		zap.L().Error("Failed to encode JSON response", zap.Error(err), zap.Int("code", code))
	}
}
