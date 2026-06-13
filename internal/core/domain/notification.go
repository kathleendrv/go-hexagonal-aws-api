package domain

// EmailNotification representa el cuerpo obligatorio que enviará la app móvil
type EmailNotification struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}