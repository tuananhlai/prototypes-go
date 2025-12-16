package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Username string
	Name     string
	Password string
}

type UserRepository struct {
	users []*User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: []*User{
			{ID: uuid.New(), Username: "johndoe", Name: "John Doe", Password: "password"},
		},
	}
}

func (u *UserRepository) GetUserByID(id uuid.UUID) (*User, error) {
	for _, user := range u.users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, fmt.Errorf("error user not found")
}

func (u *UserRepository) Login(username string, password string) (*User, error) {
	for _, user := range u.users {
		if user.Username == username && user.Password == password {
			return user, nil
		}
	}

	return nil, fmt.Errorf("error invalid credentials")
}

type RefreshToken struct {
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
}

type RefreshTokenRepository struct {
	tokens []*RefreshToken
}

func NewRefreshTokenRepository() *RefreshTokenRepository {
	return &RefreshTokenRepository{
		tokens: []*RefreshToken{},
	}
}

func (r *RefreshTokenRepository) Create(userID uuid.UUID, tokenStr string, expiredAt time.Time) *RefreshToken {
	token := &RefreshToken{
		Token:     tokenStr,
		UserID:    userID,
		ExpiresAt: expiredAt,
	}

	r.tokens = append(r.tokens, token)

	return token
}
