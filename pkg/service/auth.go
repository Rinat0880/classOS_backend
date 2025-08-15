package service

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
)

const (
	tokenTTL = 12 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) GenerateToken(username, password string) (string, error) {
	user, err := s.repo.GetUser(username, s.generatePasswordHash(password))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID,
	})

	signingKey := os.Getenv("AUTH_signingKey")
	if signingKey == "" {
		return "", fmt.Errorf("AUTH_signingKey environment variable is not set")
	}

	return token.SignedString([]byte(signingKey))
}

func (s *AuthService) CreateUser(user classosbackend.User) (int, error) {
	user.Password = s.generatePasswordHash(user.Password)

	return s.repo.CreateUser(user)
}

func (s *AuthService) generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	salt := os.Getenv("AUTH_salt")

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
