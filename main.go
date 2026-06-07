package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-hexagonal-api/internal/adapters/handler"
	"go-hexagonal-api/internal/adapters/repository"
	"go-hexagonal-api/internal/core/service"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	_ "github.com/lib/pq"
)

// Variable global para el adaptador de AWS
var lambdaAdapter *httpadapter.HandlerAdapter

func init() {
	// 1. Obtener cadena de conexión (Neon en producción, Localhost en desarrollo)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:123@localhost:5432/api_hexagonal?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("No se pudo conectar a la base de datos:", err)
	}

	// 2. Inyección de Dependencias (Arquitectura Hexagonal)
	userRepo := repository.NewPostgresRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewHttpUserHandler(userService)
	uploadHandler := handler.NewUploadHandler("./uploads")

	// 3. Definición de Rutas en el Multiplexer estándar de Go
	http.HandleFunc("/register", userHandler.Register)
	http.HandleFunc("/login", userHandler.Login)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	
	// Rutas Protegidas
	http.HandleFunc("/users", handler.JWTMiddleware(userHandler.HandleUsers))
	http.HandleFunc("/upload", handler.JWTMiddleware(uploadHandler.UploadFile))

	// 4. Inicializar el adaptador proxy para AWS Lambda usando el DefaultServeMux
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