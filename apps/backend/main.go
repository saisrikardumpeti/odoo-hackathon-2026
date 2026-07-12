package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/allocation"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/asset"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/auth"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/booking"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/category"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/department"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/employee"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/handlers/maintenance"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/middleware"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/scheduler"
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
	assetHandler := asset.NewAssetHandler(stores)
	allocationHandler := allocation.NewAllocationHandler(stores, dbPool)
	bookingHandler := booking.NewBookingHandler(stores)
	maintenanceHandler := maintenance.NewMaintenanceHandler(stores)

	schedCtx, schedCancel := context.WithCancel(context.Background())
	defer schedCancel()
	go scheduler.Start(schedCtx, stores)

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
		{
			// Read endpoints — open to all authenticated roles
			v1.GET("/assets", assetHandler.ListHandler)
			v1.GET("/assets/:id", assetHandler.GetHandler)
			v1.GET("/assets/:id/history", assetHandler.GetHistoryHandler)
			v1.GET("/categories", categoryHandler.ListHandler)

			// Resource Booking — any authenticated employee
			v1.GET("/resources/:assetId/bookings", bookingHandler.ListByResourceHandler)
			v1.POST("/bookings", bookingHandler.CreateHandler)
			v1.GET("/bookings/my", bookingHandler.ListMyBookingsHandler)
			v1.PATCH("/bookings/:id/cancel", bookingHandler.CancelHandler)
			v1.PATCH("/bookings/:id/reschedule", bookingHandler.RescheduleHandler)

			// Maintenance — any authenticated employee can create and list
			v1.POST("/maintenance", maintenanceHandler.CreateHandler)
			v1.GET("/maintenance", maintenanceHandler.ListHandler)

			// Maintenance approvals — Admin or AssetManager only
			maintenanceApproveGroup := v1.Group("")
			maintenanceApproveGroup.Use(middleware.RequireRole("Admin", "AssetManager"))
			{
				maintenanceApproveGroup.PATCH("/maintenance/:id/approve", maintenanceHandler.ApproveHandler)
				maintenanceApproveGroup.PATCH("/maintenance/:id/reject", maintenanceHandler.RejectHandler)
				maintenanceApproveGroup.PATCH("/maintenance/:id/assign-technician", maintenanceHandler.AssignTechnicianHandler)
				maintenanceApproveGroup.PATCH("/maintenance/:id/start", maintenanceHandler.StartHandler)
				maintenanceApproveGroup.PATCH("/maintenance/:id/resolve", maintenanceHandler.ResolveHandler)
			}

			// Allocation & Transfer — read open to all, allocate/return needs Admin/AssetManager/DepartmentHead
			v1.GET("/allocations/my", allocationHandler.ListMyAllocationsHandler)
			v1.GET("/allocations/overdue", allocationHandler.ListOverdueHandler)
			v1.GET("/transfers/pending", allocationHandler.ListPendingTransfersHandler)

			allocateGroup := v1.Group("")
			allocateGroup.Use(middleware.RequireRole("Admin", "AssetManager", "DepartmentHead"))
			{
				allocateGroup.POST("/allocations", allocationHandler.CreateHandler)
				allocateGroup.POST("/allocations/:id/return", allocationHandler.ReturnHandler)
			}

			// Transfer request — any authenticated employee can initiate
			v1.POST("/transfers", allocationHandler.CreateTransferHandler)

			// Transfer approval — only AssetManager or DepartmentHead
			transferApproveGroup := v1.Group("")
			transferApproveGroup.Use(middleware.RequireRole("AssetManager", "DepartmentHead"))
			{
				transferApproveGroup.PATCH("/transfers/:id/approve", allocationHandler.ApproveTransferHandler)
				transferApproveGroup.PATCH("/transfers/:id/reject", allocationHandler.RejectTransferHandler)
			}

			// Write endpoints — Admin or AssetManager
			writeGroup := v1.Group("")
			writeGroup.Use(middleware.RequireRole("Admin", "AssetManager"))
			{
				writeGroup.POST("/assets", assetHandler.RegisterHandler)
				writeGroup.POST("/assets/:id/documents", assetHandler.UploadDocumentHandler)
			}

			// Admin-only endpoints
			admin := v1.Group("")
			admin.Use(middleware.RequireRole("Admin"))
			{
				admin.GET("/departments", departmentHandler.ListHandler)
				admin.POST("/departments", departmentHandler.CreateHandler)
				admin.PATCH("/departments/:id", departmentHandler.UpdateHandler)
				admin.PATCH("/departments/:id/deactivate", departmentHandler.DeactivateHandler)

				admin.POST("/categories", categoryHandler.CreateHandler)
				admin.PATCH("/categories/:id", categoryHandler.UpdateHandler)

				admin.GET("/employees", employeeHandler.ListHandler)
				admin.PATCH("/employees/:id", employeeHandler.UpdateHandler)
				admin.PATCH("/employees/:id/role", employeeHandler.UpdateRoleHandler)
			}
		}
	}

	log.Println("Starting server on :8000...")
	router.Run(":8000")
}
