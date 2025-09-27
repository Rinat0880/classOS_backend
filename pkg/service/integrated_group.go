package service

import (
	"fmt"

	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
)

type IntegratedGroupService struct {
	repo      repository.Group
	adService *ADService
}

func NewIntegratedGroupService(repo repository.Group, adService *ADService) *IntegratedGroupService {
	return &IntegratedGroupService{
		repo:      repo,
		adService: adService,
	}
}

func (s *IntegratedGroupService) Create(checkerId int, group classosbackend.Group) (int, error) {
	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	adGroup := ADGroup{
		Name:        group.Name,
		Description: "Created by ClassOS",
	}

	err = s.adService.CreateGroup(adGroup)
	if err != nil {
		return 0, fmt.Errorf("failed to create group in AD: %w", err)
	}

	groupId, err := s.repo.CreateWithTx(tx, checkerId, group)
	if err != nil {
		s.adService.DeleteGroup(group.Name)
		return 0, fmt.Errorf("failed to create group in DB: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.adService.DeleteGroup(group.Name)
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return groupId, nil
}

func (s *IntegratedGroupService) GetAll(checkerId int) ([]classosbackend.Group, error) {
	return s.repo.GetAll(checkerId)
}

func (s *IntegratedGroupService) GetById(checkerId, groupId int) (classosbackend.Group, error) {
	return s.repo.GetById(checkerId, groupId)
}

func (s *IntegratedGroupService) Update(checkerId, groupId int, input classosbackend.UpdateGroupInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	currentGroup, err := s.repo.GetById(checkerId, groupId)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if input.Name != nil {
		adUpdates := ADGroup{
			Name: *input.Name,
		}

		err = s.adService.UpdateGroup(currentGroup.Name, adUpdates)
		if err != nil {
			return fmt.Errorf("failed to update group in AD: %w", err)
		}
	}

	err = s.repo.UpdateWithTx(tx, checkerId, groupId, input)
	if err != nil {
		return fmt.Errorf("failed to update group in DB: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *IntegratedGroupService) Delete(checkerId, groupId int) error {
	group, err := s.repo.GetById(checkerId, groupId)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = s.adService.DeleteGroup(group.Name)
	if err != nil {
		return fmt.Errorf("failed to delete group from AD: %w", err)
	}

	err = s.repo.DeleteWithTx(tx, checkerId, groupId)
	if err != nil {
		return fmt.Errorf("failed to delete group from DB: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
