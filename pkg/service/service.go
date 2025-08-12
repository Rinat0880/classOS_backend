package service

import "github.com/rinat0880/classOS_backend/pkg/repository"

type Authorization interface {
}

type Group interface {
}

type User interface {
}

type Service struct {
	Authorization
	Group
	User
}

func NewService(repos *repository.Repository) *Service {
	return &Service{}
}
