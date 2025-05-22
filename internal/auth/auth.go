package auth

import (
	"context"
	"encoding/base64"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

const jwtExpirationTime = time.Minute * 20 // expiration

func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-=_!@#$%^&*()")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// accepts JWT or Basic Auth
// a JWT token is send through Authorization-Jwt header when Basic Auth is successful for further authentication
func Auth(jwtSecretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for JWT token in the custom Authorization-Jwt header
			jwtHeader := r.Header.Get("Authorization-Jwt")
			if strings.HasPrefix(jwtHeader, "Bearer ") {
				claims := &jwt.StandardClaims{}

				// Parse and verify the JWT token
				token, err := jwt.ParseWithClaims(jwtHeader[len("Bearer "):], claims, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecretKey), nil
				})

				if err == nil && token.Valid {
					// Valid token; add username to context and proceed
					ctx := context.WithValue(r.Context(), UserContextKey, claims.Subject)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Fall back to Basic Auth if no JWT token is provided or if token is invalid
			basicAuthHeader := r.Header.Get("Authorization")
			if basicAuthHeader == "" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized: Please provide Basic Auth credentials", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(basicAuthHeader, "Basic ") {
				http.Error(w, "Unauthorized: Invalid authorization scheme", http.StatusUnauthorized)
				return
			}

			// Decode the Basic Auth credentials
			payload, err := base64.StdEncoding.DecodeString(basicAuthHeader[len("Basic "):])
			if err != nil {
				http.Error(w, "Unauthorized: Invalid base64 encoding", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(string(payload), ":", 2)
			if len(parts) != 2 {
				http.Error(w, "Unauthorized: Invalid credentials format", http.StatusUnauthorized)
				return
			}
			username, password := parts[0], parts[1]

			// Check credentials using bcrypt
			if ok, err := checkCredentials(username, password); !ok || err != nil {
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
				return
			}

			// Create JWT token if Basic Auth is successful
			claims := &jwt.StandardClaims{
				Subject:   username,
				ExpiresAt: time.Now().Add(jwtExpirationTime).Unix(),
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(jwtSecretKey))
			if err != nil {
				http.Error(w, "Could not create token", http.StatusInternalServerError)
				return
			}

			// Respond with the JWT token in the Authorization-Jwt header
			w.Header().Set("Authorization-Jwt", "Bearer "+tokenString)

			// Proceed to next handler, passing username in context
			ctx := context.WithValue(r.Context(), UserContextKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
