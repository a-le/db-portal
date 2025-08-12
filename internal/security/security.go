package security

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"db-portal/internal/contextkeys"

	"github.com/golang-jwt/jwt/v5"
)

// JWTSecretKey is the secret key used for signing JWT tokens.
var JWTSecretKey []byte

// JWTMiddleware is a middleware that checks for a valid JWT token in the request header.
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if after, ok := strings.CutPrefix(tokenString, "Bearer "); ok {
			tokenString = after
		}

		if tokenString == "" {
			// If not found, try form value "jwt"
			if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				r.ParseMultipartForm(10 << 20)
				tokenString = r.FormValue("jwt")
			}
		}

		if tokenString == "" {
			http.Error(w, "Unauthorized, no token found", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return JWTSecretKey, nil
		}, jwt.WithValidMethods([]string{"HS256"}))

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized, token not valid", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized, can't parse claims", http.StatusUnauthorized)
			return
		}

		username, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Unauthorized, can't get username from claim", http.StatusUnauthorized)
			return
		}

		ctx := contextkeys.SetUsername(r.Context(), username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoadJWTSecretKey(path string) ([]byte, error) {
	key, err := os.ReadFile(path)
	if err != nil || len(key) == 0 {
		return nil, err
	}
	return key, nil
}

func GenerateJWTSecretKey() ([]byte, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(b)), nil
}

func SaveJWTSecretKey(path string, key []byte) error {
	return os.WriteFile(path, key, 0600)
}
