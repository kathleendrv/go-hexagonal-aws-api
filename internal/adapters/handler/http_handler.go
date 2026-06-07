package handler

import (
	"encoding/json"
	"fmt"
	"go-hexagonal-api/internal/core/ports"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type HttpUserHandler struct {
	service ports.UserService
}

func NewHttpUserHandler(service ports.UserService) *HttpUserHandler {
	return &HttpUserHandler{service: service}
}

// Middleware de Autenticación JWT
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Se requiere token de autenticación", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("tu_clave_secreta_super_segura"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (h *HttpUserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(req.Name, req.Email, req.Password)
// Dentro de tu método Register en el Handler de Go:
if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    // 💡 IMPORTANTE: Esto le mandará el mensaje real (ej. "relation 'users' does not exist") a Flutter
    json.NewEncoder(w).Encode(map[string]string{
        "error": fmt.Sprintf("Error en el núcleo de Go: %v", err),
    })
    return
}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *HttpUserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *HttpUserHandler) HandleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.URL.Query().Get("id")

	switch r.Method {
	case http.MethodGet:
		if idStr != "" {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			user, err := h.service.GetUser(id)
			if err != nil {
				http.Error(w, "Usuario no encontrado", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(user)
		} else {
			users, _ := h.service.ListUsers()
			json.NewEncoder(w).Encode(users)
		}

	case http.MethodPut:
		id, _ := strconv.ParseInt(idStr, 10, 64)
		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		user, err := h.service.UpdateUser(id, req.Name, req.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(user)

	case http.MethodDelete:
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if err := h.service.DeleteUser(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Método no soportado", http.StatusMethodNotAllowed)
	}
}