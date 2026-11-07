package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"
)

func Bearer(token string, next http.Handler) http.Handler {

	expected := sha256.Sum256([]byte(token))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		headers := r.Header.Values("Authorization")

		if len(headers) != 1 {
			unauthorized(w)
			return
		}

		scheme, supplied, ok := strings.Cut(headers[0], " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") {
			unauthorized(w)
			return
		}

		supplied = strings.TrimSpace(supplied)

		if supplied == "" ||
			strings.ContainsAny(supplied, " \t\r\n") {
			unauthorized(w)
			return
		}

		actual := sha256.Sum256([]byte(supplied))

		if subtle.ConstantTimeCompare(expected[:], actual[:]) != 1 {
			unauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
