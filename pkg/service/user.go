package service

import (
	"crypto/sha1"
	"fmt"

	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
)

type UserService struct {
	repo      repository.User
	groupRepo repository.Group
}

func NewUserService(repo repository.User, groupRepo repository.Group) *UserService {
	return &UserService{repo: repo, groupRepo: groupRepo}
}

func (s *UserService) Create(checkerId, groupId int, user classosbackend.User) (int, error) {
	_, err := s.groupRepo.GetById(checkerId, groupId)
	if err != nil {
		return 0, err
	}

	user.Password = s.generatePasswordHash(user.Password)

	return s.repo.Create(groupId, user)
}


func (s *UserService) generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	salt := getSalt()
	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func (s *UserService) GetAll(checkerId int) ([]classosbackend.User, error) {
	return s.repo.GetAll(checkerId)
}

func (s *UserService) GetById(checkerId, user_id int) (classosbackend.User, error) {
	return s.repo.GetById(checkerId, user_id)
}

func (s *UserService) Delete(checkerId, user_id int) error {
	return s.repo.Delete(checkerId, user_id)
}

func (s *UserService) Update(checkerId, user_id int, input classosbackend.UpdateUserInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	if input.Password != nil {
        hashedPassword := s.generatePasswordHash(*input.Password)
        input.Password = &hashedPassword
    }
	
	return s.repo.Update(checkerId, user_id, input)
}