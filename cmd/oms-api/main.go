package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/keyurKalariya/OMS/cmd/oms-api/handlers"
	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	pb "github.com/keyurKalariya/OMS/cmd/oms-api/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDB() (*gorm.DB, error) {
	// Get database configuration from environment variables, with defaults for local development
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "oms")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Retry database connection (useful in Docker when DB might not be ready immediately)
	maxRetries := 30
	retryDelay := 2 * time.Second
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			log.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if _, err := sqlDB.Exec(`CREATE SCHEMA IF NOT EXISTS grpc`); err != nil {
		return nil, err
	}

	if _, err := sqlDB.Exec("SET search_path TO grpc"); err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Item{}, &models.User{}, &models.Order{}, &models.OrderItem{}, &models.UserOrder{})
	if err != nil {
		return nil, err
	}

	log.Println("Connected to the PostgreSQL database using GORM v2")
	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Initialize the database
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	log.Println("Database connection initialized successfully")
	log.Print("<=========================================================>")
	log.Print("<==================== Starting OMS ====================>")
	log.Print("<=========================================================>")

	grpcPort := ":" + getEnv("GRPC_PORT", "8089")
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer()

	// Enable gRPC reflection
	reflection.Register(grpcServer)

	// Register services
	omsItemService := &handlers.OmsItemServiceServer{DB: db}
	pb.RegisterOmsItemServiceServer(grpcServer, omsItemService)

	omsUserService := &handlers.OmsUserServiceServer{DB: db}
	pb.RegisterUserServiceServer(grpcServer, omsUserService)

	omsOrderService := &handlers.OrderServiceServer{DB: db}
	pb.RegisterOrderServiceServer(grpcServer, omsOrderService)

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("Starting gRPC server on port %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %s", err)
		}
	}()

	// Start grpcui in a goroutine (optional, can be disabled via environment variable)
	if getEnv("ENABLE_GRPCUI", "true") == "true" {
		go func() {
			grpcuiPort := getEnv("GRPCUI_PORT", "8080")
			grpcHost := getEnv("GRPC_HOST", "localhost")
			// Bind to 0.0.0.0 to make it accessible from outside the container
			log.Printf("Starting grpcui on http://0.0.0.0:%s", grpcuiPort)
			grpcuiCmd := exec.Command("grpcui", "-plaintext", "-bind", "0.0.0.0", "-port", grpcuiPort, grpcHost+grpcPort)
			grpcuiCmd.Stdout = os.Stdout
			grpcuiCmd.Stderr = os.Stderr

			if err := grpcuiCmd.Start(); err != nil {
				log.Printf("Failed to start grpcui: %v (continuing without grpcui)", err)
				return
			}

			// Wait for grpcui to finish
			if err := grpcuiCmd.Wait(); err != nil {
				log.Printf("grpcui process exited: %v", err)
			}
		}()
	}

	// Block the main function so that the program doesn't exit
	select {}

}
