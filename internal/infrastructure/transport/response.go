package transport

import (
	"encoding/json"
	"net/http"
)

type WebResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJson(w http.ResponseWriter, code int, message string, data interface{}) {
	response := WebResponse{Code: code, Message: message, Data: data}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
