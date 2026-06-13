package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"GO-HEXAGONAL-API/internal/core/domain" // Asegúrate de que coincida con tu módulo de go.mod

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type NotificationHandler struct {
	snsClient *sns.SNS
	topicArn  string
}

func NewNotificationHandler() *NotificationHandler {
	// Inicializamos la sesión nativa de AWS dentro de la Lambda
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	
	return &NotificationHandler{
		snsClient: sns.New(sess),
		topicArn:  os.Getenv("SNS_TOPIC_ARN"), // Terraform inyectará esta variable automáticamente
	}
}

func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	var req domain.EmailNotification
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cuerpo de solicitud inválido"})
		return
	}

	// Validación Obligatoria de QA
	if req.Email == "" || req.Subject == "" || req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Todos los campos (email, subject, message) son obligatorios"})
		return
	}

	// Convertimos la petición a string JSON para meterla en el mensaje de SNS
	payloadBytes, _ := json.Marshal(req)

	log.Printf("Publicando mensaje en SNS Topic: %s", h.topicArn)

	// Publicar directo al Tópico de SNS
	_, err = h.snsClient.Publish(&sns.PublishInput{
		Message:  aws.String(string(payloadBytes)),
		TopicArn: aws.String(h.topicArn),
	})

	if err != nil {
		log.Printf("Error publicando en SNS: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error interno al procesar notificación en AWS"})
		return
	}

	log.Println("Mensaje distribuido con éxito a SNS")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Mensaje enviado correctamente"})
}