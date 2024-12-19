package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os"
// 	"os/exec"
// 	"strings"

// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/keyurKalariya/OMS/cmd/oms-api/handlers"
// 	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
// 	pb "github.com/keyurKalariya/OMS/cmd/oms-api/protobuf"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/metadata"
// 	"google.golang.org/grpc/reflection"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// // JWT secret key
// var jwtSecret = []byte("your-secret-key")

// // JWT Authentication Interceptor
// func jwtAuthInterceptor(
// 	ctx context.Context,
// 	req interface{},
// 	info *grpc.UnaryServerInfo,
// 	handler grpc.UnaryHandler,
// ) (interface{}, error) {
// 	// Extract token from metadata
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok {
// 		return nil, fmt.Errorf("missing metadata")
// 	}

// 	token := ""
// 	if vals, ok := md["authorization"]; ok && len(vals) > 0 {
// 		// Expect "Bearer <token>"
// 		authHeader := strings.SplitN(vals[0], " ", 2)
// 		if len(authHeader) == 2 && authHeader[0] == "Bearer" {
// 			token = authHeader[1]
// 		}
// 	}

// 	if token == "" {
// 		return nil, fmt.Errorf("missing or invalid token")
// 	}

// 	// Parse and validate the JWT token
// 	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
// 		// Ensure the token method is HMAC
// 		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
// 		}
// 		return jwtSecret, nil
// 	})

// 	if err != nil || !parsedToken.Valid {
// 		return nil, fmt.Errorf("invalid token: %v", err)
// 	}

// 	// Proceed with the request
// 	return handler(ctx, req)
// }

// func initDB() (*gorm.DB, error) {
// 	connStr := "host=localhost user=root password=root dbname=oms sslmode=disable"
// 	// connStr := "host=postgres-container-40 user=root password=root dbname=oms sslmode=disable"

// 	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	sqlDB, err := db.DB()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if _, err := sqlDB.Exec(`CREATE SCHEMA IF NOT EXISTS grpc`); err != nil {
// 		return nil, err
// 	}

// 	if _, err := sqlDB.Exec("SET search_path TO grpc"); err != nil {
// 		return nil, err
// 	}

// 	if err := sqlDB.Ping(); err != nil {
// 		return nil, err
// 	}

// 	err = db.AutoMigrate(&models.Item{}, &models.User{}, &models.Order{}, &models.OrderItem{}, &models.UserOrder{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Println("Connected to the PostgreSQL database using GORM v2")
// 	return db, nil
// }

// func main() {
// 	// Initialize the database
// 	db, err := initDB()
// 	if err != nil {
// 		log.Fatalf("Failed to connect to the database: %v", err)
// 	}

// 	grpcPort := ":8089"
// 	lis, err := net.Listen("tcp", grpcPort)
// 	if err != nil {
// 		log.Fatalf("Failed to create listener: %s", err)
// 	}

// 	// Add JWT authentication interceptor to gRPC server
// 	grpcServer := grpc.NewServer(
// 		grpc.UnaryInterceptor(jwtAuthInterceptor),
// 	)

// 	// Enable gRPC reflection
// 	reflection.Register(grpcServer)

// 	// Register services
// 	omsItemService := &handlers.OmsItemServiceServer{DB: db}
// 	pb.RegisterOmsItemServiceServer(grpcServer, omsItemService)

// 	omsUserService := &handlers.OmsUserServiceServer{DB: db}
// 	pb.RegisterUserServiceServer(grpcServer, omsUserService)

// 	omsOrderService := &handlers.OrderServiceServer{DB: db}
// 	pb.RegisterOrderServiceServer(grpcServer, omsOrderService)

// 	// Start gRPC server in a goroutine
// 	go func() {
// 		log.Printf("Starting gRPC server on port %s", grpcPort)
// 		if err := grpcServer.Serve(lis); err != nil {
// 			log.Fatalf("Failed to serve gRPC server: %s", err)
// 		}
// 	}()

// 	// Start grpcui in a goroutine
// 	go func() {
// 		// log.Println("Starting grpcui on http://localhost:8080")
// 		grpcuiCmd := exec.Command("grpcui", "-plaintext", "localhost"+grpcPort)
// 		grpcuiCmd.Stdout = os.Stdout
// 		grpcuiCmd.Stderr = os.Stderr

// 		if err := grpcuiCmd.Start(); err != nil {
// 			log.Fatalf("Failed to start grpcui: %v", err)
// 		}

// 		// Wait for grpcui to finish
// 		if err := grpcuiCmd.Wait(); err != nil {
// 			log.Printf("grpcui process exited: %v", err)
// 		}
// 	}()

// 	select {}
// }
