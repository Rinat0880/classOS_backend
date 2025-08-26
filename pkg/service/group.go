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

func (s *GroupService) Create(checkerId int, group classosbackend.Group) (int, error) {
	return s.repo.Create(checkerId, group)
}

func (s *GroupService) GetAll(checkerId int) ([]classosbackend.Group, error) {
	return s.repo.GetAll(checkerId)
}

func (s *GroupService) GetById(checkerId, groupId int) (classosbackend.Group, error) {
	return s.repo.GetById(checkerId, groupId)
}

func (s *GroupService) Delete(checkerId, groupId int) error {
	return s.repo.Delete(checkerId, groupId) 
}

func (s *GroupService) Update(checkerId, groupId int, input classosbackend.UpdateGroupInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	return s.repo.Update(checkerId, groupId, input)
}