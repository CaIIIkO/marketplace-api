package main

import (
	"log"
	_ "marketplace-api/docs"
	"marketplace-api/internal/advertisement"
	"marketplace-api/internal/auth"
	"marketplace-api/internal/db"
	"marketplace-api/internal/user"
	"net/http"
	"os"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Marketplace API
// @version 1.0
// @description API для маркетплейса — регистрация, авторизация и объявления (Тестовое задание для https://internship.vk.company/vacancy/1146)

// @contact.name Chirkov Alexandr
// @contact.url https://t.me/CAIIIKO0
// @contact.email chirkov.a.y@yandex.ru

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey AuthToken
// @in header
// @name Authorization
// @description Введите JWT токен с префиксом Bearer

// @schemes http

func main() {
	log.Println("marketplace-api is starting...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "supersecretjwtkey"
	}

	dsn := os.Getenv("DATABASE_DSN")

	//Подключение к базе данных
	pool := db.Connect(dsn)

	//Инциализация jwtManager
	jwtManager := auth.NewJWTManager(secretKey)

	//инциализация хэндлеров и сервисов
	userRepo := user.NewRepository(pool)
	userService := user.NewUserService(userRepo, jwtManager)
	userHandler := user.NewUserHandler(userService)

	adRepo := advertisement.NewAdRepository(pool)
	adService := advertisement.NewAdService(adRepo)
	adHandler := advertisement.NewAdHandler(adService)

	//http
	mux := http.NewServeMux()

	mux.HandleFunc("/register", userHandler.Register) //POST
	mux.HandleFunc("/login", userHandler.Login)       //POST

	mux.Handle("/advertisement", auth.AuthMiddleware(jwtManager, http.HandlerFunc(adHandler.CreateAd)))        //POST
	mux.Handle("/advertisement/", auth.OptionalAuthMiddleware(jwtManager, http.HandlerFunc(adHandler.ListAd))) //GET

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	//Swagger
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+port+"/swagger/doc.json"),
	))

	// Запуск сервера
	log.Printf("server is running http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
