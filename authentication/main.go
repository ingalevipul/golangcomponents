package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Use a secure secret key in production!
var secretKey = []byte("your-super-secret-key")

// CustomClaims defines the structure of the data inside the token
type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJWT(username string) (string, error) {
	// Set custom and standard claims
	claims := CustomClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func VerifyJWT(tokenString string) (*CustomClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	// Extract claims if token is valid
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// Login Handler (Simulated)
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// In a real app, you would verify username and password here
	username := "john_doe"

	token, err := GenerateJWT(username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("Your Token: %s", token)))
}

// Protected Handler
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header (Format: Bearer <token>)
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		http.Error(w, "Unauthorized: Missing or invalid token format", http.StatusUnauthorized)
		return
	}

	tokenString := authHeader[7:]

	// Verify the token
	claims, err := VerifyJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// If valid, grant access
	w.Write([]byte(fmt.Sprintf("Welcome to the secret club, %s!", claims.Username)))
}

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract token from Authorization header (Format: Bearer <token>)
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			http.Error(w, "Unauthorized: Missing or invalid token format", http.StatusUnauthorized)
			return // Stop execution here
		}

		tokenString := authHeader[7:]

		// 2. Verify the token
		claims, err := VerifyJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return // Stop execution here
		}

		// 3. (Optional) Pass user info down to the actual handler using context
		// This makes claims.Username accessible inside your protected routes
		ctx := r.Context()
		// We use a custom type for context keys to avoid collisions
		type contextKey string
		r = r.WithContext(context.WithValue(ctx, contextKey("username"), claims.Username))

		// 4. Token is valid! Call the next handler in the chain
		next(w, r)
	}
}

func main() {
	server := http.NewServeMux()
	// Change .Handle to .HandleFunc
	server.HandleFunc("GET /login", loginHandler)
	server.HandleFunc("GET /somepage", JWTMiddleware(
		func(w http.ResponseWriter, r *http.Request) {
			response := UserResponse{
				ID:    1,
				Name:  "Alice Smith",
				Email: "alice@example.com",
			}

			// 3. Set the Content-Type header to application/json
			w.Header().Set("Content-Type", "application/json")

			// 4. Set the status code (optional, defaults to 200 OK)
			w.WriteHeader(http.StatusOK)

			// 5. Use json.NewEncoder to stream the JSON directly to the response writer
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}))
	http.ListenAndServe("localhost:8080", server)
}
