package security

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/csrf"

	"db-portal/internal/internaldb"

	"math/rand"

	"golang.org/x/crypto/bcrypt"
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

// SecurityConfig holds all security-related middleware configuration
type SecurityConfig struct {
	Store        *internaldb.Store
	JWTSecretKey string
	CORSOptions  cors.Options
}

// NewSecurityConfig creates default security configuration
func NewSecurityConfig(store *internaldb.Store, jwtKey string, allowedOrigins []string) *SecurityConfig {
	return &SecurityConfig{
		Store:        store,
		JWTSecretKey: jwtKey,
		CORSOptions: cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Authorization"},
			AllowCredentials: true,
			MaxAge:           300,
		},
	}
}

func (c *SecurityConfig) SetupSecurityMiddleware(r *chi.Mux) {
	// Security headers
	r.Use(middleware.SetHeader("X-Content-Type-Options", "nosniff"))
	r.Use(middleware.SetHeader("X-Frame-Options", "deny"))
	r.Use(middleware.SetHeader("X-XSS-Protection", "1; mode=block"))

	// CORS
	r.Use(cors.Handler(c.CORSOptions))

	// CSRF protection
	csrfMiddleware := csrf.Protect(
		[]byte(c.JWTSecretKey),
		csrf.Path("/"),
		csrf.Secure(true),
		csrf.CookieName("csrf"),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.MaxAge(int(jwtExpirationTime.Seconds())),
		csrf.FieldName("_csrf"),
	)

	// Add CSRF middleware and token header in one place
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First apply CSRF protection
			handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Then add CSRF token to response headers
				w.Header().Set("X-CSRF-Token", csrf.Token(r))
				next.ServeHTTP(w, r)
			}))
			handler.ServeHTTP(w, r)
		})
	})
}

// Auth is a middleware that handles authentication using JWT or Basic Auth
func (c *SecurityConfig) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for JWT token in standard Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(authHeader[len("Bearer "):], claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(c.JWTSecretKey), nil
			})
			if err == nil && token.Valid {
				ctx := context.WithValue(r.Context(), UserContextKey, claims.Subject)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Check JWT from cookie as fallback
		if cookie, err := r.Cookie("jwt"); err == nil {
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(c.JWTSecretKey), nil
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
		if ok, err := c.Store.CheckUserCredentials(username, password); !ok || err != nil {
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
		tokenString, err := token.SignedString([]byte(c.JWTSecretKey))
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
