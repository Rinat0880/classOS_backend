package service

import (
	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
)

type Authorization interface {
	CreateUser(user classosbackend.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type Group interface {
	Create(userId int, group classosbackend.Group) (int, error)
	GetAll(userId int) ([]classosbackend.Group, error)
	GetById(userId, groupId int) (classosbackend.Group, error)
}

type User interface {
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
	}
}
