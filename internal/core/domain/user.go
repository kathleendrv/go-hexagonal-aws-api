package domain

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"` // omitempty para no devolverla en las respuestas
	CreatedAt time.Time `json:"created_at"`
}