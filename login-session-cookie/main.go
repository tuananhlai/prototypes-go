package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	addr              = ":8080"
	userContextKey    = "user"
	sessionCookieName = "session_id"
)

// Demonstrate the session cookie login flow.
//
// 1. Run `http :8080/login username="johndoe" password="password"`
// 2. Take note of the session ID cookie.
// 3. Run `http :8080/me Cookie:session_id=0d8d05cf-3938-4329-bf8b-473eaa154c49` to fetch user information.
func main() {
	mux := http.NewServeMux()

	userRepo := NewUserRepository()
	sessionRepo := NewSessionRepository()
	authMiddleware := NewAuthMiddleware(userRepo, sessionRepo)

	mux.HandleFunc("POST /login", loginHandler(userRepo, sessionRepo))
	// TODO: test this endpoint
	mux.HandleFunc("POST /logout", logoutHandler(sessionRepo))
	mux.Handle("GET /me", authMiddleware.Wrap(meHandler()))

	log.Println("starting server at", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start server")
	}
}

type MeResponseDTO struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func logoutHandler(sessionRepo *SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionIDCookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		sessionRepo.Delete(sessionIDCookie.Value)
		deleteCookie(w, sessionCookieName)
	}
}

// meHandler is a HTTP endpoint that returns user information.
func meHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := GetRequestUser(r)
		if err != nil {
			http.Error(w, "can not extract user info", http.StatusInternalServerError)
			return
		}

		res := MeResponseDTO{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

type LoginRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponseDTO struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func loginHandler(userRepo *UserRepository, sessionRepo *SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginReq LoginRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			http.Error(w, "can not decode request body", http.StatusBadRequest)
			return
		}

		user, err := userRepo.Login(loginReq.Username, loginReq.Password)
		if err != nil {
			http.Error(w, fmt.Sprintf("can not login: %v", err), http.StatusBadRequest)
			return
		}

		sessionID := sessionRepo.Create(user.ID)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		res := LoginResponseDTO{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
		}
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func CreateRequestWithUser(base *http.Request, user *User) *http.Request {
	return base.WithContext(context.WithValue(base.Context(), userContextKey, user))
}

func GetRequestUser(r *http.Request) (*User, error) {
	rawVal := r.Context().Value(userContextKey)
	user, ok := rawVal.(*User)
	if !ok {
		return nil, fmt.Errorf("error user not found or invalid")
	}

	return user, nil
}

type AuthMiddleware struct {
	userRepo    *UserRepository
	sessionRepo *SessionRepository
}

func NewAuthMiddleware(userRepo *UserRepository, sessionRepo *SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (am *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, err := am.sessionRepo.GetUserID(sessionID.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		user, err := am.userRepo.GetUserByID(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, CreateRequestWithUser(r, user))
	})
}

func deleteCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}
