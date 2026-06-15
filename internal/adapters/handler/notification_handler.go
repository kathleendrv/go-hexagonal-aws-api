package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"go-hexagonal-api/internal/core/domain"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type NotificationHandler struct {
	snsClient *sns.SNS
	topicArn  string
}

func NewNotificationHandler() *NotificationHandler {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	
	return &NotificationHandler{
		snsClient: sns.New(sess),
		topicArn:  os.Getenv("SNS_TOPIC_ARN"),
	}
}

func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al enviar mensaje."})
		return
	}

	var req domain.EmailNotification
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al enviar mensaje."})
		return
	}

	// Validación de campos obligatorios
	if req.Email == "" || req.Subject == "" || req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al enviar mensaje."})
		return
	}

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
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al enviar mensaje."})
		return
	}

	log.Println("Mensaje distribuido con éxito a SNS")
	w.WriteHeader(http.StatusOK)
	// Respuesta exacta solicitada por la asignación para Flutter
	json.NewEncoder(w).Encode(map[string]string{"message": "Mensaje enviado correctamente."})
}