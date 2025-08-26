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
	Create(userId int, group classosbackend.Group) (int, error)
	GetAll(userId int) ([]classosbackend.Group, error)
	GetById(userId, groupId int) (classosbackend.Group, error)
	Delete(userId, groupId int) error
	Update(userId, groupId int, input classosbackend.UpdateGroupInput) error
}

type User interface {
	Create(groupId int, user classosbackend.User) (int, error)
	GetAll(userId, groupId int) ([]classosbackend.User, error)
}

type Repository struct {
	Authorization
	Group
	User
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Group: NewGroupPostgres(db),
		User: NewUserPostgres(db),
	}
}
