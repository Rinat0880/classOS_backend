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
	Create(checkerId int, group classosbackend.Group) (int, error)
	GetAll(checkerId int) ([]classosbackend.Group, error)
	GetById(checkerId, groupId int) (classosbackend.Group, error)
	Delete(checkerId, groupId int) error
	Update(checkerId, groupId int, input classosbackend.UpdateGroupInput) error
}

type User interface {
	Create(groupId int, user classosbackend.User) (int, error)
	GetAll(checkerId, groupId int) ([]classosbackend.User, error)
	GetById(checkerId, user_id int) (classosbackend.User, error)
	Delete(checkerId, user_id int) error
	Update(checkerId, user_id int, input classosbackend.UpdateUserInput) error
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
