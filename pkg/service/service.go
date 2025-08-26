package service

import (
	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
)

type Authorization interface {
	CreateUser(user classosbackend.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, string, error)
}

type Group interface {
	Create(checkerId int, group classosbackend.Group) (int, error)
	GetAll(checkerId int) ([]classosbackend.Group, error)
	GetById(checkerId, groupId int) (classosbackend.Group, error)
	Delete(checkerId, groupId int) error
	Update(checkerId, groupId int, input classosbackend.UpdateGroupInput) error
}

type User interface {
	Create(checkerId, groupId int, user classosbackend.User) (int, error)
	GetAll(checkerId, groupId int) ([]classosbackend.User, error)
	GetById(checkerId, user_id int) (classosbackend.User, error)
	Delete(checkerId, user_id int) error
	Update(checkerId, user_id int, input classosbackend.UpdateUserInput) error
}

type Service struct {
	Authorization
	Group
	User
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		Group: NewGroupService(repos.Group),
		User: NewUserService(repos.User, repos.Group),
	}
}
