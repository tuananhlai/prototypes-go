package main

import (
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
	addr            = ":8080"
	keyFilePath     = "key.pem"
	tokenCookieName = "token"
)

func main() {
	mux := http.NewServeMux()

	privateKey, err := loadECDSAPrivateKey(keyFilePath)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := NewUserRepository()
	tokenFactory := NewAuthTokenService(privateKey)

	mux.HandleFunc("POST /login", loginHandler(userRepo, tokenFactory))

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
		})

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(loginResDTO)
	}
}

// AuthTokenService creates and verifies JWT tokens for authentication.
type AuthTokenService struct {
	key           *ecdsa.PrivateKey
	signingMethod jwt.SigningMethod
}

func NewAuthTokenService(key *ecdsa.PrivateKey) *AuthTokenService {
	return &AuthTokenService{key: key, signingMethod: jwt.SigningMethodES256}
}

func (atg *AuthTokenService) Create(userID int, expiration time.Duration) (string, error) {
	t := jwt.NewWithClaims(atg.signingMethod, jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
	})

	tokenStr, err := t.SignedString(atg.key)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
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
