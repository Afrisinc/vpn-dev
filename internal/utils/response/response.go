package response

import (
	"encoding/json"
	"net/http"
)

// JSONResponse is the standard response structure
type JSONResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message,omitempty"`
	Data    interface{}          `json:"data,omitempty"`
	Error   *ErrorDetail         `json:"error,omitempty"`
	RequestID string              `json:"requestId,omitempty"`
}

// ErrorDetail contains structured error information
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// sendResponse is a helper function to send JSON responses
func sendResponse(w http.ResponseWriter, statusCode int, resp JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// OK sends a 200 OK response with data
func OK(w http.ResponseWriter, message string, data interface{}) {
	sendResponse(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, message string, data interface{}) {
	sendResponse(w, http.StatusCreated, JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Success is an alias for OK (legacy compatibility)
func Success(w http.ResponseWriter, message string, data interface{}) {
	OK(w, message, data)
}

// Error sends an error response with a specific status code
func Error(w http.ResponseWriter, statusCode int, message string, err interface{}) {
	var errorDetail *ErrorDetail
	if err != nil {
		if ed, ok := err.(*ErrorDetail); ok {
			errorDetail = ed
		} else if str, ok := err.(string); ok {
			errorDetail = &ErrorDetail{
				Code:    "error",
				Message: str,
			}
		}
	}

	sendResponse(w, statusCode, JSONResponse{
		Success: false,
		Message: message,
		Error:   errorDetail,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string, fields map[string]string) {
	sendResponse(w, http.StatusBadRequest, JSONResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "bad_request",
			Message: message,
			Fields:  fields,
		},
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	sendResponse(w, http.StatusUnauthorized, JSONResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "unauthorized",
			Message: message,
		},
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	sendResponse(w, http.StatusForbidden, JSONResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "forbidden",
			Message: message,
		},
	})
}

// NotFound sends a 404 Not Found response
func NotFound(w http.ResponseWriter, message string) {
	sendResponse(w, http.StatusNotFound, JSONResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "not_found",
			Message: message,
		},
	})
}

// Conflict sends a 409 Conflict response
func Conflict(w http.ResponseWriter, message string) {
	sendResponse(w, http.StatusConflict, JSONResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    "conflict",
			Message: message,
		},
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(w http.ResponseWriter, message string, code string) {
	sendResponse(w, http.StatusInternalServerError, JSONResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(w http.ResponseWriter) {
	sendResponse(w, http.StatusTooManyRequests, JSONResponse{
		Success: false,
		Message: "Too many requests",
		Error: &ErrorDetail{
			Code:    "rate_limit_exceeded",
			Message: "Too many requests. Please try again later.",
		},
	})
}
