package middleware

import (
	"net/http"
)

func AuthMiddleWare(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
