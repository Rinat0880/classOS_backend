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
	Create(userId int, group classosbackend.Group) (int, error)
	GetAll(userId int) ([]classosbackend.Group, error)
	GetById(userId, groupId int) (classosbackend.Group, error)
	Delete(userId, groupId int) error
	Update(userId, groupId int, input classosbackend.UpdateGroupInput) error
}

type User interface {
	Create(userId, groupId int, user classosbackend.User) (int, error)
	GetAll(userId, groupId int) ([]classosbackend.User, error)
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
