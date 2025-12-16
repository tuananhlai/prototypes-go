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
	users []*User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: []*User{
			{ID: 1, Username: "johndoe", Name: "John Doe", Password: "password"},
		},
	}
}

func (u *UserRepository) GetUserByID(id int) (*User, error) {
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
