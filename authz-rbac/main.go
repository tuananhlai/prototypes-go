package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 1. Define RBAC Types
type Role string
type Permission string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"

	PermRead  Permission = "read:data"
	PermWrite Permission = "write:data"
)

var rolePermissions = map[Role]map[Permission]bool{
	RoleAdmin:  {PermRead: true, PermWrite: true},
	RoleEditor: {PermRead: true, PermWrite: true},
	RoleViewer: {PermRead: true},
}

var jwtKey = []byte("my_secret_key") // In production, use an environment variable

// CustomClaims defines the structure of the JWT payload
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	mux := http.NewServeMux()

	// Public: Login to get a token
	mux.HandleFunc("POST /login", handleLogin)

	// Protected: Must have a valid JWT AND the correct permission
	mux.Handle("GET /data", withAuth(enforce(PermRead, handleRead)))
	mux.Handle("POST /data", withAuth(enforce(PermWrite, handleWrite)))

	addr := ":8080"
	log.Printf("RBAC + JWT server running on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// handleLogin simulates authentication and returns a JWT
func handleLogin(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	var role Role

	// Mock user database
	switch userID {
	case "alice":
		role = RoleAdmin
	case "bob":
		role = RoleEditor
	case "charlie":
		role = RoleViewer
	default:
		http.Error(w, "Invalid user. Try ?user=alice, bob, or charlie", http.StatusUnauthorized)
		return
	}

	// Create claims
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Login successful for %s (%s)\nToken: %s\n", userID, role, tokenString)
}

// withAuth middleware validates the JWT and injects claims into context
func withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// enforce middleware checks if the Role in the JWT claims has the required Permission
func enforce(perm Permission, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("claims").(*CustomClaims)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !hasPermission(claims.Role, perm) {
			log.Printf("FORBIDDEN: User %s (Role: %s) tried to access %s", claims.UserID, claims.Role, perm)
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Forbidden: Role '%s' lacks '%s' permission\n", claims.Role, perm)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hasPermission(role Role, perm Permission) bool {
	perms, ok := rolePermissions[role]
	return ok && perms[perm]
}

func handleRead(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*CustomClaims)
	fmt.Fprintf(w, "Success! User: %s, Role: %s, Action: READ\n", claims.UserID, claims.Role)
}

func handleWrite(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*CustomClaims)
	fmt.Fprintf(w, "Success! User: %s, Role: %s, Action: WRITE\n", claims.UserID, claims.Role)
}
