package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	addr             = ":8080"
	keyFilePath      = "key.pem"
	tokenCookieName  = "token"
	userIDContextKey = "user"
)

func main() {
	mux := http.NewServeMux()

	privateKey, err := loadECDSAPrivateKey(keyFilePath)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := NewUserRepository()
	authTokenService := NewAuthTokenService(privateKey)
	authMiddleware := NewAuthMiddleware(userRepo, authTokenService)

	mux.HandleFunc("POST /login", loginHandler(userRepo, authTokenService))
	mux.Handle("GET /me", authMiddleware.Wrap(meHandler(userRepo)))

	log.Println("starting server on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
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

func loginHandler(userRepo *UserRepository, tokenFactory *AuthTokenService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginReqDTO LoginRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&loginReqDTO); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := userRepo.Login(loginReqDTO.Username, loginReqDTO.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		oneWeek := time.Hour * 24 * 7
		token, err := tokenFactory.Create(user.ID, oneWeek)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		loginResDTO := LoginResponseDTO{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
		}

		http.SetCookie(w, &http.Cookie{
			Name:     tokenCookieName,
			Value:    token,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		json.NewEncoder(w).Encode(loginResDTO)
	}
}

type MeResponseDTO struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func meHandler(userRepo *UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetRequestUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		meResDTO := MeResponseDTO{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
		}

		json.NewEncoder(w).Encode(meResDTO)
	}
}

// AuthTokenService creates and verifies JWT tokens for authentication.
type AuthTokenService struct {
	privateKey    *ecdsa.PrivateKey
	signingMethod jwt.SigningMethod
}

func NewAuthTokenService(key *ecdsa.PrivateKey) *AuthTokenService {
	return &AuthTokenService{
		privateKey:    key,
		signingMethod: jwt.SigningMethodES256,
	}
}

func (atg *AuthTokenService) Create(userID int, expiration time.Duration) (string, error) {
	t := jwt.NewWithClaims(atg.signingMethod, jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
	})

	tokenStr, err := t.SignedString(atg.privateKey)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// Validate verifies the structure and signature of the given JWT token, then parse
// and return the `subject`.
func (atg *AuthTokenService) Validate(tokenStr string) (userID int, err error) {
	parser := jwt.NewParser(jwt.WithExpirationRequired())
	token, err := parser.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		// TODO: which kind of logic is expected here?

		// The expected return types are (*ecdsa.PublicKey, error). I'm still not sure
		// why though.
		return &atg.privateKey.PublicKey, nil
	})
	if err != nil {
		return
	}

	userIDStr, err := token.Claims.GetSubject()
	if err != nil {
		return
	}

	return strconv.Atoi(userIDStr)
}

// loadECDSAPrivateKey reads an ECDSA private key from a PEM file
// from the given path.
func loadECDSAPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	rawData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading key from file: %v", err)
	}

	key, err := jwt.ParseECPrivateKeyFromPEM(rawData)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %v", err)
	}

	return key, nil
}

type AuthMiddleware struct {
	userRepo         *UserRepository
	authTokenService *AuthTokenService
}

func NewAuthMiddleware(userRepo *UserRepository, authTokenService *AuthTokenService) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo:         userRepo,
		authTokenService: authTokenService,
	}
}

func (am *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(tokenCookieName)
		if err != nil {
			http.Error(w, "token not found: "+err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := am.authTokenService.Validate(tokenCookie.Value)
		if err != nil {
			http.Error(w, "invalid token received: "+err.Error(), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, CreateRequestWithUserID(r, userID))
	})
}

func CreateRequestWithUserID(base *http.Request, userID int) *http.Request {
	return base.WithContext(context.WithValue(base.Context(), userIDContextKey, userID))
}

func GetRequestUserID(r *http.Request) (int, error) {
	rawVal := r.Context().Value(userIDContextKey)
	user, ok := rawVal.(int)
	if !ok {
		return 0, fmt.Errorf("error user ID not found or invalid")
	}

	return user, nil
}
