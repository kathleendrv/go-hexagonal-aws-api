package repository

import (
	"database/sql"
	"go-hexagonal-api/internal/core/domain"
	"go-hexagonal-api/internal/core/ports"

	_ "github.com/lib/pq"
)

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) ports.UserRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(user *domain.User) error {
	query := `INSERT INTO users (name, email, password, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRow(query, user.Name, user.Email, user.Password, user.CreatedAt).Scan(&user.ID)
}

func (r *postgresRepository) GetByID(id int64) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *postgresRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *postgresRepository) Update(user *domain.User) error {
	query := `UPDATE users SET name = $1, email = $2 WHERE id = $3`
	_, err := r.db.Exec(query, user.Name, user.Email, user.ID)
	return err
}

func (r *postgresRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *postgresRepository) GetAll() ([]domain.User, error) {
	query := `SELECT id, name, email, created_at FROM users`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}