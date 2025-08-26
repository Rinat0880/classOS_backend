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

func (s *UserService) Create(userId, groupId int, user classosbackend.User) (int, error) {
	_, err := s.groupRepo.GetById(userId, groupId)
	if err != nil {
		return 0, err
	}

	user.Password = s.generatePasswordHash(user.Password)

	return s.repo.Create(groupId, user)
}

func (s *UserService) GetAll(userId, groupId int) ([]classosbackend.User, error) {
	return s.repo.GetAll(userId, groupId)
}

func (s *UserService) generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	salt := getSalt()

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
