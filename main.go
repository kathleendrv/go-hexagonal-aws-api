package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"go-hexagonal-api/internal/adapters/handler"
	"go-hexagonal-api/internal/adapters/repository"
	"go-hexagonal-api/internal/core/service"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	_ "github.com/lib/pq"
)

// Variable global para el adaptador de AWS
var lambdaAdapter *httpadapter.HandlerAdapter

func init() {
	log.Println("Iniciando Lambda v2 - Comprobando conexión...")

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:123@localhost:5432/api_hexagonal?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("ERROR en sql.Open: %v", err)
		panic(err)
	}

	// Forzar el ping y capturar el error real
	if err = db.Ping(); err != nil {
		log.Printf("ERROR de conexión a Postgres (Neon): %v", err)
		panic(fmt.Sprintf("Fallo crítico de base de datos: %v", err)) 
	}

	log.Println("Conexión a base de datos exitosa.")

	// Inyección de Dependencias
	userRepo := repository.NewPostgresRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewHttpUserHandler(userService)
	uploadHandler := handler.NewUploadHandler("./uploads")

	// nicializamos primero el Handler de Notificaciones SNS
	notifHandler := handler.NewNotificationHandler()

	// Rutas
	http.HandleFunc("/register", userHandler.Register)
	http.HandleFunc("/login", userHandler.Login)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	http.HandleFunc("/users", handler.JWTMiddleware(userHandler.HandleUsers))
	http.HandleFunc("/upload", handler.JWTMiddleware(uploadHandler.UploadFile))
	
	// Ahora sí usamos el notifHandler correctamente inicializado arriba
	http.HandleFunc("/notifications/send", notifHandler.SendNotification)
	
	lambdaAdapter = httpadapter.New(http.DefaultServeMux)
}

// Handler que interactúa directamente con AWS API Gateway
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return lambdaAdapter.ProxyWithContext(ctx, req)
}

func main() {
	// Si existe esta variable, significa que estamos ejecutándonos dentro de AWS Lambda
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(Handler)
	} else {
		// Si no existe, estamos en nuestra computadora local (Docker o Go run)
		fmt.Println("Servidor corriendo localmente en http://0.0.0.0:8080")
		log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
	}
}