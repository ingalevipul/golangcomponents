package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"tutproj/hashlib"
	"tutproj/redisdb"
)

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var secretKey []byte

type responsetoken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type Status string

const (
	accesstoken  string = "access token"
	refreshtoken string = "refresh token"
)

var demoUsers = map[string]string{
	"alice": "password123",
}

func verifyCredentials(username, password string) bool {
	expected, ok := demoUsers[username]
	return ok && expected == password
}

func refreshTokenKey(token string) string {
	return hashlib.HashJWT(token)
}

// CustomClaims defines the structure of the data inside the token
type CustomClaims struct {
	Username  string `json:"username"`
	TokenType Status

	jwt.RegisteredClaims
}

func GenerateJWT(username string, tokentype Status) (string, error) {
	var expirytime *jwt.NumericDate
	issuedat := jwt.NewNumericDate(time.Now())
	if tokentype == Status(accesstoken) {
		expirytime = jwt.NewNumericDate(time.Now().Add(15 * 60 * time.Second))
		issuedat = jwt.NewNumericDate(time.Now())
	} else {
		expirytime = jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour))
		issuedat = jwt.NewNumericDate(time.Now())
	}
	claims := CustomClaims{
		Username:  username,
		TokenType: Status(tokentype),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expirytime,
			IssuedAt:  issuedat,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyJWT(tokenString string, tokentype Status) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if tokentype == Status(accesstoken) {
		if claims.TokenType != Status(accesstoken) {
			return nil, fmt.Errorf("invalid token")
		}
	} else if claims.TokenType != Status(refreshtoken) {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func loginHandler(rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}
	if !verifyCredentials(req.Username, req.Password) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	accessToken, err := GenerateJWT(req.Username, Status(accesstoken))
	if err != nil {
		http.Error(w, "Failed to generate Access token", http.StatusInternalServerError)
		return
	}
	newRefreshToken, err := GenerateJWT(req.Username, Status(refreshtoken))
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}
	response := responsetoken{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}

	ctx := r.Context()
	if err := rdb.Set(ctx, refreshTokenKey(newRefreshToken), "true", 7*24*time.Hour).Err(); err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		http.Error(w, "Unauthorized: Missing or invalid token format", http.StatusUnauthorized)
		return
	}

	tokenString := authHeader[7:]

	claims, err := VerifyJWT(tokenString, Status(accesstoken))
	if err != nil {
		log.Printf("JWT verification failed: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Write([]byte(fmt.Sprintf("Welcome to the secret club, %s!", claims.Username)))
}

func refreshHandler(rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.RefreshToken == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	oldKey := refreshTokenKey(req.RefreshToken)

	val, err := rdb.GetDel(ctx, oldKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if val != "true" {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	claims, err := VerifyJWT(req.RefreshToken, Status(refreshtoken))
	if err != nil {
		log.Printf("refresh token JWT invalid after Redis hit: %v", err)
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	newAccessToken, err := GenerateJWT(claims.Username, Status(accesstoken))
	if err != nil {
		http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
		return
	}
	newRefreshToken, err := GenerateJWT(claims.Username, Status(refreshtoken))
	if err != nil {
		http.Error(w, "Failed to generate new refresh token", http.StatusInternalServerError)
		return
	}

	if err := rdb.Set(ctx, refreshTokenKey(newRefreshToken), "true", 7*24*time.Hour).Err(); err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responsetoken{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			http.Error(w, "Unauthorized: Missing or invalid token format", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]

		claims, err := VerifyJWT(tokenString, Status(accesstoken))
		if err != nil {
			log.Printf("JWT verification failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		type contextKey string
		r = r.WithContext(context.WithValue(ctx, contextKey("username"), claims.Username))

		next(w, r)
	}
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	secretKey = []byte(secret)

	rdb := redisdb.CreateConnection()

	server := http.NewServeMux()
	server.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(rdb, w, r)
	})

	server.HandleFunc("POST /refresh", func(w http.ResponseWriter, r *http.Request) {
		refreshHandler(rdb, w, r)
	})
	server.HandleFunc("GET /somepage", JWTMiddleware(
		func(w http.ResponseWriter, r *http.Request) {
			response := UserResponse{
				ID:    1,
				Name:  "Alice Smith",
				Email: "alice@example.com",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}))
	http.ListenAndServe("localhost:8080", server)
}
