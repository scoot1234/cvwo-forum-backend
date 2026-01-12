package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"CVWO-Backend/controllers"
	"CVWO-Backend/db"
	"CVWO-Backend/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	gdb, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}

	if err := gdb.AutoMigrate(
		&models.Topic{},
		&models.User{},
		&models.Post{},
		&models.Comment{},
	); err != nil {
		log.Fatalf("db migrate error: %v", err)
	}

	if err := db.SeedDefaultTopics(gdb); err != nil {
	log.Fatalf("seed topics error: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	
	r.Use(cors.Handler(cors.Options{
	  AllowedOrigins: []string{
	    "https://cvwo-forum-frontend-xyb2.onrender.com",
	    "http://localhost:5173",
	  },
	  AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
	  AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	  AllowCredentials: false,
	  MaxAge: 300,
	}))

	authController := controllers.NewAuthController(gdb)
	commentsController := controllers.NewCommentsController(gdb)
	postsController := controllers.NewPostsController(gdb)
	topicsController := controllers.NewTopicsController(gdb)

	authController.RegisterRoutes(r)
	commentsController.RegisterRoutes(r)
	postsController.RegisterRoutes(r)
	topicsController.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	fmt.Printf("Server running on port %s\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
