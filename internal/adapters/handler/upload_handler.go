package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type UploadHandler struct {
	UploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	// Asegurar que la carpeta exista al inicializar
	os.MkdirAll(uploadDir, os.ModePerm)
	return &UploadHandler{UploadDir: uploadDir}
}

func (h *UploadHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parsear multipart form (límite 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Archivo demasiado grande", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error al recuperar el archivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Crear ruta del archivo destino
	filePath := filepath.Join(h.UploadDir, handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error al guardar localmente", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error al escribir el archivo", http.StatusInternalServerError)
		return
	}

	// Construir respuesta con la URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": "http://localhost:8080/uploads/" + handler.Filename,
	})
}