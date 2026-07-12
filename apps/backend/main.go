package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/auth"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/category"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/department"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/employee"
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
	departmentHandler := department.NewDepartmentHandler(stores)
	categoryHandler := category.NewCategoryHandler(stores)
	employeeHandler := employee.NewEmployeeHandler(stores)

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

		v1 := api.Group("/v1")
		v1.Use(middleware.AuthRequired())
		v1.Use(middleware.RequireRole("Admin"))
		{
			v1.GET("/departments", departmentHandler.ListHandler)
			v1.POST("/departments", departmentHandler.CreateHandler)
			v1.PATCH("/departments/:id", departmentHandler.UpdateHandler)
			v1.PATCH("/departments/:id/deactivate", departmentHandler.DeactivateHandler)

			v1.GET("/categories", categoryHandler.ListHandler)
			v1.POST("/categories", categoryHandler.CreateHandler)
			v1.PATCH("/categories/:id", categoryHandler.UpdateHandler)

			v1.GET("/employees", employeeHandler.ListHandler)
			v1.PATCH("/employees/:id", employeeHandler.UpdateHandler)
			v1.PATCH("/employees/:id/role", employeeHandler.UpdateRoleHandler)
		}
	}

	log.Println("Starting server on :8000...")
	router.Run(":8000")
}
