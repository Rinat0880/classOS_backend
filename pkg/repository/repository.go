package repository

import (
	"github.com/jmoiron/sqlx"
	classosbackend "github.com/rinat0880/classOS_backend"
)

type Authorization interface {
	CreateUser(user classosbackend.User) (int, error)
	GetUser(username, password string) (classosbackend.User, error)
}

type Group interface {
}

type User interface {
}

type Repository struct {
	Authorization
	Group
	User
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
	}
}
