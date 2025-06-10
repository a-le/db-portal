//go:build ignore
// +build ignore

package auth

import (
	"context"
	"encoding/base64"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"db-portal/internal/internaldb"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/csrf"
)

const jwtExpirationTime = time.Minute * 20 // expiration

type contextKey string

const UserContextKey = contextKey("username")

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

// WithCSRFToken is middleware that adds the CSRF token to response headers
func WithCSRFToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CSRF-Token", csrf.Token(r))
		next.ServeHTTP(w, r)
	})
}

// NewCSRFMiddleware creates a new CSRF protection middleware
func NewCSRFMiddleware(key string) func(http.Handler) http.Handler {
	return csrf.Protect(
		[]byte(key),
		csrf.Path("/"),
		csrf.Secure(true),
		csrf.CookieName("csrf"),
		csrf.HttpOnly(true),
	)
}

// Auth is a middleware that handles authentication using JWT or Basic Auth.
// It first checks for a JWT token in the Authorization header or cookie.
// If no valid JWT is found, it falls back to Basic Auth.
// If Basic Auth is successful, it generates a JWT token and sets it in the Authorization header and a secure HTTP-only cookie.
// The JWT token is valid for a limited time and is used for subsequent requests.
func Auth(connService *internaldb.Store, jwtSecretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for JWT token in standard Authorization header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				claims := &jwt.RegisteredClaims{}

				// Parse and verify the JWT token
				token, err := jwt.ParseWithClaims(authHeader[len("Bearer "):], claims, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecretKey), nil
				})
				if err == nil && token.Valid {
					// Valid token; add username to context and proceed
					ctx := context.WithValue(r.Context(), UserContextKey, claims.Subject)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Check JWT from cookie as fallback
			if cookie, err := r.Cookie("jwt"); err == nil {
				claims := &jwt.RegisteredClaims{}
				token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecretKey), nil
				})
				if err == nil && token.Valid {
					ctx := context.WithValue(r.Context(), UserContextKey, claims.Subject)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Fall back to Basic Auth if no valid JWT found
			basicAuthHeader := r.Header.Get("Authorization")
			if basicAuthHeader == "" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized: Please provide credentials", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(basicAuthHeader, "Basic ") {
				http.Error(w, "Unauthorized: Invalid authorization scheme", http.StatusUnauthorized)
				return
			}

			// Decode Basic Auth credentials
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

			// Validate credentials
			if ok, err := connService.CheckUserCredentials(username, password); !ok || err != nil {
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
				return
			}

			// Create JWT token after successful Basic Auth
			claims := &jwt.RegisteredClaims{
				Subject:   username,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpirationTime)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(jwtSecretKey))
			if err != nil {
				http.Error(w, "Could not create token", http.StatusInternalServerError)
				return
			}

			// Set standard Authorization header
			w.Header().Set("Authorization", "Bearer "+tokenString)

			// Set secure HTTP-only cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "jwt",
				Value:    tokenString,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   int(jwtExpirationTime.Seconds()),
			})

			// Proceed with request
			ctx := context.WithValue(r.Context(), UserContextKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// accepts JWT or Basic Auth
// a JWT token is send through Authorization-Jwt header when Basic Auth is successful for further authentication
// func Auth(connService *internaldb.Store, jwtSecretKey string) func(next http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			// Check for JWT token in the custom Authorization-Jwt header
// 			jwtHeader := r.Header.Get("Authorization-Jwt")
// 			if strings.HasPrefix(jwtHeader, "Bearer ") {
// 				claims := &jwt.RegisteredClaims{}

// 				// Parse and verify the JWT token
// 				token, err := jwt.ParseWithClaims(jwtHeader[len("Bearer "):], claims, func(token *jwt.Token) (interface{}, error) {
// 					return []byte(jwtSecretKey), nil
// 				})
// 				if err == nil && token.Valid {
// 					// Valid token; add username to context and proceed
// 					ctx := context.WithValue(r.Context(), UserContextKey, claims.Subject)
// 					next.ServeHTTP(w, r.WithContext(ctx))
// 					return
// 				}
// 			}

// 			// Fall back to Basic Auth if no JWT token is provided or if token is invalid
// 			basicAuthHeader := r.Header.Get("Authorization")
// 			if basicAuthHeader == "" {
// 				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
// 				http.Error(w, "Unauthorized: Please provide Basic Auth credentials", http.StatusUnauthorized)
// 				return
// 			}

// 			if !strings.HasPrefix(basicAuthHeader, "Basic ") {
// 				http.Error(w, "Unauthorized: Invalid authorization scheme", http.StatusUnauthorized)
// 				return
// 			}

// 			// Decode the Basic Auth credentials
// 			payload, err := base64.StdEncoding.DecodeString(basicAuthHeader[len("Basic "):])
// 			if err != nil {
// 				http.Error(w, "Unauthorized: Invalid base64 encoding", http.StatusUnauthorized)
// 				return
// 			}

// 			parts := strings.SplitN(string(payload), ":", 2)
// 			if len(parts) != 2 {
// 				http.Error(w, "Unauthorized: Invalid credentials format", http.StatusUnauthorized)
// 				return
// 			}
// 			username, password := parts[0], parts[1]

// 			// Check credentials using bcrypt
// 			if ok, err := connService.CheckUserCredentials(username, password); !ok || err != nil {
// 				if err != nil {
// 					http.Error(w, err.Error(), http.StatusInternalServerError)
// 					return
// 				}
// 				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
// 				http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
// 				return
// 			}

// 			// Create JWT token if Basic Auth is successful
// 			claims := &jwt.RegisteredClaims{
// 				Subject:   username,
// 				ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpirationTime)),
// 			}

// 			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 			tokenString, err := token.SignedString([]byte(jwtSecretKey))
// 			if err != nil {
// 				http.Error(w, "Could not create token", http.StatusInternalServerError)
// 				return
// 			}

// 			// Respond with the JWT token in the Authorization-Jwt header
// 			w.Header().Set("Authorization-Jwt", "Bearer "+tokenString)

// 			// Proceed to next handler, passing username in context
// 			ctx := context.WithValue(r.Context(), UserContextKey, username)
// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }
