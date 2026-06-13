package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Estructura del mensaje que vendrá desde SNS dentro de SQS
type SNSMessage struct {
	Message string `json:"Message"`
}

// Estructura del correo original enviado por Flutter
type EmailPayload struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Printf("[Notification-Lambda] Iniciando procesamiento de %d mensajes desde SQS", len(sqsEvent.Records))

	for _, record := range sqsEvent.Records {
		log.Printf("Procesando SQS MessageID: %s", record.MessageId)

		// 1. SQS envuelve el mensaje de SNS. Primero extraemos el cuerpo de SNS.
		var snsMsg SNSMessage
		err := json.Unmarshal([]byte(record.Body), &snsMsg)
		if err != nil {
			log.Printf(" Error al decodificar estructura SNS: %v", err)
			continue
		}

		// 2. Ahora decodificamos el Payload del correo original enviado por el usuario
		var emailData EmailPayload
		err = json.Unmarshal([]byte(snsMsg.Message), &emailData)
		if err != nil {
			log.Printf(" Error al decodificar Payload de correo: %v", err)
			continue
		}

		// 3. Procesar y simular/enviar el correo electrónico
		log.Println("-------------------------------------------------------------")
		log.Printf("ENVIANDO CORREO ELECTRÓNICO EXITOSAMENTE")
		log.Printf("Para: %s", emailData.Email)
		log.Printf("Asunto: %s", emailData.Subject)
		log.Printf("Mensaje: %s", emailData.Message)
		log.Println("-------------------------------------------------------------")
		log.Printf(" Evento registrado con éxito en CloudWatch para ID: %s", record.MessageId)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}