package service

import (
	"fmt"
	"os"

	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
	"github.com/sirupsen/logrus"
)

type IntegratedUserService struct {
	repo        repository.User
	groupRepo   repository.Group
	authService *AuthService
	adService   *ADService
}

var groupname string

func NewIntegratedUserService(repo repository.User, groupRepo repository.Group, authService *AuthService, adService *ADService) *IntegratedUserService {
	return &IntegratedUserService{
		repo:        repo,
		groupRepo:   groupRepo,
		authService: authService,
		adService:   adService,
	}
}

func (s *IntegratedUserService) Create(checkerId, groupId int, user classosbackend.User) (int, error) {
	_, err := s.groupRepo.GetById(checkerId, groupId)
	if err != nil {
		return 0, err
	}

	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	adUser := s.convertUserToADUser(user)

	if user.GroupName != nil {
		groupname = *user.GroupName
		logrus.WithField("groupname", groupname).Info("obtained successfullay groupname")
	} else{
		return 0, fmt.Errorf("failed to obtain groupname for AD")
	}

	err = s.adService.CreateUser(adUser, user.Password, groupname)
	if err != nil {
		return 0, fmt.Errorf("failed to create user in AD: %w", err)
	}

	user.Password = s.authService.GeneratePasswordHash(user.Password)
	userId, err := s.repo.CreateWithTx(tx, groupId, user)
	if err != nil {
		s.adService.DeleteUser(user.Username)
		return 0, fmt.Errorf("failed to create user in DB: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.adService.DeleteUser(user.Username)
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return userId, nil
}

func (s *IntegratedUserService) GetAll(checkerId int) ([]classosbackend.User, error) {
	return s.repo.GetAll(checkerId)
}

func (s *IntegratedUserService) GetById(checkerId, userId int) (classosbackend.User, error) {
	return s.repo.GetById(checkerId, userId)
}

func (s *IntegratedUserService) Update(checkerId, userId int, input classosbackend.UpdateUserInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	currentUser, err := s.repo.GetById(checkerId, userId)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if input.Name != nil || input.Username != nil {
		adUpdates := ADUser{}
		if input.Name != nil {
			adUpdates.DisplayName = *input.Name
		}
		if input.Username != nil {
			adUpdates.SamAccountName = *input.Username
		}

		if input.GroupName != nil {
			groupname = *input.GroupName
			logrus.WithField("groupname", groupname).Info("obtained successfullay groupname")
		} else{
			return fmt.Errorf("failed to obtain groupname for AD")
		}

		err = s.adService.UpdateUser(currentUser.Username, adUpdates, groupname)
		if err != nil {
			return fmt.Errorf("failed to update user in AD: %w", err)
		}
	}

	if input.Password != nil {
		err = s.adService.ChangeUserPassword(currentUser.Username, *input.Password)
		if err != nil {
			return fmt.Errorf("failed to change password in AD: %w", err)
		}

		hashedPassword := s.authService.GeneratePasswordHash(*input.Password)
		input.Password = &hashedPassword
	}

	err = s.repo.UpdateWithTx(tx, checkerId, userId, input)
	if err != nil {
		return fmt.Errorf("failed to update user in DB: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *IntegratedUserService) Delete(checkerId, userId int) error {
	user, err := s.repo.GetById(checkerId, userId)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.Username == "admin01" {
		return fmt.Errorf("cannot delete super admin")
	}

	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = s.adService.DeleteUser(user.Username)
	if err != nil {
		return fmt.Errorf("failed to delete user from AD: %w", err)
	}

	err = s.repo.DeleteWithTx(tx, checkerId, userId)
	if err != nil {
		return fmt.Errorf("failed to delete user from DB: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *IntegratedUserService) SyncAllFromAD() error {
	return s.adService.SyncAllUsersFromAD()
}

func (s *IntegratedUserService) ValidateADConnection() error {
	return s.adService.TestConnection()
}

func (s *IntegratedUserService) convertUserToADUser(user classosbackend.User) ADUser {
	return ADUser{
		SamAccountName: user.Username,
		DisplayName:    user.Name,
		EmailAddress:   user.Username + "@" + os.Getenv("AD_DOMAIN"),
		Enabled:        true,
	}
}
