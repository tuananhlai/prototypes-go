package main

import (
	"fmt"

	"github.com/google/uuid"
)

type SessionRepository struct {
	sessions map[string]int
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]int),
	}
}

func (s *SessionRepository) CreateSession(userID int) string {
	sessionID := uuid.New().String()
	s.sessions[sessionID] = userID

	return sessionID
}

func (s *SessionRepository) GetUserID(sessionID string) (int, error) {
	userID, ok := s.sessions[sessionID]
	if !ok {
		return 0, fmt.Errorf("error invalid session ID")
	}

	return userID, nil
}

type User struct {
	ID       int
	Username string
	Name     string
	Password string
}

type UserRepository struct {
	users map[int]*User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: map[int]*User{
			1: {ID: 1, Username: "johndoe", Name: "John Doe", Password: "password"},
		},
	}
}

func (u *UserRepository) GetUserByID(id int) (*User, error) {
	user, ok := u.users[id]
	if !ok {
		return nil, fmt.Errorf("error user not found")
	}

	return user, nil
}

func (u *UserRepository) Login(username string, password string) (*User, error) {
	for _, user := range u.users {
		if user.Username == username && user.Password == password {
			return user, nil
		}
	}

	return nil, fmt.Errorf("error invalid credentials")
}
