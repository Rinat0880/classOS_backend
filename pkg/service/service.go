package service

import (
	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
)

type Authorization interface {
	CreateUser(user classosbackend.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, string, error)
	GeneratePasswordHash(password string) string
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
	GetAll(checkerId int) ([]classosbackend.User, error)
	GetById(checkerId, userId int) (classosbackend.User, error)
	Delete(checkerId, userId int) error
	Update(checkerId, userId int, input classosbackend.UpdateUserInput) error
}

type Service struct {
	Authorization
	Group
	User
}

func NewService(repos *repository.Repository) *Service {
	adService := NewADService()
	authService := NewAuthService(repos.Authorization)

	return &Service{
		Authorization: authService,
		Group:         NewIntegratedGroupService(repos.Group, adService),
		User:          NewIntegratedUserService(repos.User, repos.Group, authService, adService),
	}
}
