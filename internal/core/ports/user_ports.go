package ports

import "go-hexagonal-api/internal/core/domain"

// UserRepository es el puerto de salida (hacia la base de datos)
type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id int64) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id int64) error
	GetAll() ([]domain.User, error)
}

// UserService es el puerto de entrada (hacia los handlers HTTP)
type UserService interface {
	Register(name, email, password string) (*domain.User, error)
	Login(email, password string) (string, error) // Devuelve el token JWT
	GetUser(id int64) (*domain.User, error)
	UpdateUser(id int64, name, email string) (*domain.User, error)
	DeleteUser(id int64) error
	ListUsers() ([]domain.User, error)
}