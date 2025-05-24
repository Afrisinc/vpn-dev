package response

import (
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func Success(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, statusCode int, message string, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(JSONResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}
