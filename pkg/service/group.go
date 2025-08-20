package service

import (
	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
)

type GroupService struct {
	repo repository.Group
}

func NewGroupService(repo repository.Group) *GroupService {
	return &GroupService{repo: repo}
}

func (s *GroupService) Create(userId int, group classosbackend.Group) (int, error) {
	return s.repo.Create(userId, group)
}

func (s *GroupService) GetAll(userId int) ([]classosbackend.Group, error) {
	return s.repo.GetAll(userId)
}

func (s *GroupService) GetById(userId, groupId int) (classosbackend.Group, error) {
	return s.repo.GetById(userId, groupId)
}