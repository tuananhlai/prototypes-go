package main

import (
	"fmt"
)

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
