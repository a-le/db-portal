package auth

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const jwtExpirationTime = time.Minute * 20 // expiration
const UserContextKey = contextKey("username")

type File struct {
	modTime time.Time
	content []string
}

var htpasswdFile = File{
	modTime: time.Time{},
	content: []string{},
}

// CheckCredentials checks if the provided username and password match the stored credentials in the file.
func checkCredentials(filePath, username, password string) (bool, error) {

	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	if !info.ModTime().Equal(htpasswdFile.modTime) {
		fmt.Printf("htpasswd file %s loaded\n", filePath)
		file, err := os.Open(filePath)
		if err != nil {
			return false, err
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return false, fmt.Errorf("error reading file: %w", err)
		}

		htpasswdFile.content = lines
		htpasswdFile.modTime = info.ModTime()
	}

	for lineNumber, line := range htpasswdFile.content {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("malformed line. Line number: %d", lineNumber)
		}

		storedUsername := parts[0]
		storedHash := parts[1]

		if storedUsername == username {
			if !checkPasswordHash(password, storedHash) {
				return false, nil // Password does not match
			}
			return true, nil // Username and password match
		}
	}

	return false, nil // Username not found
}

func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// accepts JWT or Basic Auth
// a JWT token is send through Authorization-Jwt header when Basic Auth is successful for further authentication
func Auth(htpasswdPath string, jwtSecretKey string) func(next http.Handler) http.Handler {
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
			if ok, err := checkCredentials(htpasswdPath, username, password); !ok || err != nil {
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
