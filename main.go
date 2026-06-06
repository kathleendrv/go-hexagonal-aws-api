package main

import (
	"database/sql"
	"fmt"
	"go-hexagonal-api/internal/adapters/handler"
	"go-hexagonal-api/internal/adapters/repository"
	"go-hexagonal-api/internal/core/service"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // Asegúrate de tener el driver importado
)

func main() {
	// MODIFICACIÓN DOCKER: Si existe la variable de entorno, usa esa cadena (para Docker), si no, usa localhost (desarrollo local).
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:123@localhost:5432/api_hexagonal?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("No se pudo conectar a la base de datos:", err)
	}

	// Inyección de Dependencias
	userRepo := repository.NewPostgresRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewHttpUserHandler(userService)
	
	// Nuevo Handler de subidas
	uploadHandler := handler.NewUploadHandler("./uploads")

	// Definición de Rutas Públicas
	http.HandleFunc("/register", userHandler.Register)
	http.HandleFunc("/login", userHandler.Login)

	// Ruta para servir las imágenes en el navegador/app (Preview de Flutter)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Rutas Protegidas por JWT Middleware (Añadimos /upload aquí)
	http.HandleFunc("/users", handler.JWTMiddleware(userHandler.HandleUsers))
	http.HandleFunc("/upload", handler.JWTMiddleware(uploadHandler.UploadFile))

	fmt.Println("Servidor corriendo en http://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}