package service

import (
	"errors"
	"go-hexagonal-api/internal/core/domain"
	"go-hexagonal-api/internal/core/ports"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("tu_clave_secreta_super_segura") // Cambiar en producción

type userService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) ports.UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(name, email, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}
	user.Password = "" // Ocultar
	return user, nil
}

func (s *userService) Login(email, password string) (string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// Generar Token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtKey)
}

func (s *userService) GetUser(id int64) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) UpdateUser(id int64, name, email string) (*domain.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	user.Name = name
	user.Email = email

	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) DeleteUser(id int64) error {
	return s.repo.Delete(id)
}

func (s *userService) ListUsers() ([]domain.User, error) {
	return s.repo.GetAll()
}