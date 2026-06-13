package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type UploadHandler struct {
	UploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	// Ya no creamos carpetas locales (os.MkdirAll) porque en AWS Lambda el almacenamiento es efímero.
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

	// ADAPTACIÓN PARA AWS LAMBDA (SERVERLESS):
	log.Printf(" Archivo recibido con éxito en la Lambda: %s (%d bytes)", handler.Filename, handler.Size)

	// Construir respuesta con una URL simulada en la nube
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": "https://go-hexagonal-uploads.s3.amazonaws.com/" + handler.Filename,
	})
}