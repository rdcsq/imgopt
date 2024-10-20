package middleware

import (
	"encoding/json"
	"imgopt/models"
	"net/http"
	"strings"
)

type AuthWrapper struct {
	AllowedTokens []models.AllowedToken
}

func (aw *AuthWrapper) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.Split(r.Header.Get("Authorization"), " ")
		if len(token) != 2 || token[0] != "Bearer" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "Bad Request"})
			return
		}

		for _, allowedToken := range aw.AllowedTokens {
			if token[1] == allowedToken.Token {
				r.Header.Set("X-Identity", allowedToken.Alias)
				next.ServeHTTP(w, r)
				return
			}
		}

		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
	})
}
