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
	addr        = ":8080"
	keyFilePath = "key.pem"
)

func main() {
	mux := http.NewServeMux()

	privateKey, err := loadECDSAPrivateKey(keyFilePath)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := NewUserRepository()
	tokenFactory := NewAuthTokenFactory(privateKey)

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
	Token    string `json:"token"`
}

func loginHandler(userRepo *UserRepository, tokenFactory *AuthTokenFactory) http.HandlerFunc {
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
			Token:    token,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(loginResDTO)
	}
}

type AuthTokenFactory struct {
	key *ecdsa.PrivateKey
}

func NewAuthTokenFactory(key *ecdsa.PrivateKey) *AuthTokenFactory {
	return &AuthTokenFactory{key: key}
}

func (atg *AuthTokenFactory) Create(userID int, expiration time.Duration) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
	})

	tokenStr, err := t.SignedString(atg.key)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

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
