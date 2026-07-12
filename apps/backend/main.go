package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/auth"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/middleware"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/seed"
)

func initDatabase() (*pgxpool.Pool, error) {
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		log.Fatalf("DATABASE_URL not set in environment")
	}

	pool, err := pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to the database!")
	return pool, nil
}

func main() {
	dbPool, err := initDatabase()
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}
	defer dbPool.Close()

	if err := seed.RunMigrations(context.Background(), dbPool); err != nil {
		log.Fatalf("Could not initialize database schema: %v", err)
	}

	stores := repository.NewStorageRegistry(dbPool)
	authHandler := auth.NewAuthHandler(stores)

	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/auth/signup", authHandler.SignupHandler)
		api.POST("/auth/login", authHandler.LoginHandler)
		api.POST("/auth/refresh", authHandler.RefreshTokenHandler)
		api.POST("/auth/forgot-password", authHandler.ForgotPasswordHandler)

		auth := api.Group("/auth")
		auth.Use(middleware.AuthRequired())
		{
			auth.GET("/me", authHandler.MeHandler)
		}
	}

	log.Println("Starting server on :8000...")
	router.Run(":8000")
}
