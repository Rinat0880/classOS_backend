package service

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/repository"
)

const tokenTTL = 12 * time.Hour

var ErrInvalidCredentials = errors.New("incorrect login or password")

type tokenClaims struct {
	jwt.StandardClaims
	CheckerId int    `json:"checker_id"`
	Role      string `json:"role"`
}

type AuthService struct {
	repo repository.Authorization
}

func getSigningKey() string {
	return os.Getenv("AUTH_signingKey")
}

func getSalt() string {
	return os.Getenv("AUTH_salt")
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) GenerateToken(username, password string) (string, error) {
	user, err := s.repo.GetUser(username, s.GeneratePasswordHash(password))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("auth.GenerateToken: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		CheckerId: user.ID,
		Role:      user.Role,
	})

	signingKey := getSigningKey()
	if signingKey == "" {
		return "", fmt.Errorf("AUTH_signingKey environment variable is not set")
	}

	return token.SignedString([]byte(signingKey))
}

func (s *AuthService) CreateUser(user classosbackend.User) (int, error) {
	user.Password = s.GeneratePasswordHash(user.Password)
	return s.repo.CreateUser(user)
}

func (s *AuthService) ParseToken(accessToken string) (int, string, error) {
	signingKey := getSigningKey()
	if signingKey == "" {
		return 0, "", fmt.Errorf("AUTH_signingKey env var is not set")
	}

	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		return 0, "", err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, "", errors.New("token claims are not type of *tokenClaims")
	}

	return claims.CheckerId, claims.Role, nil
}

func (s *AuthService) GeneratePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	salt := getSalt()
	result := fmt.Sprintf("%x", hash.Sum([]byte(salt)))
	return result
}
