package repository

import "github.com/jmoiron/sqlx"

type Authorization interface {
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
	return &Repository{}
}
