package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	pb "github.com/keyurKalariya/OMS/cmd/oms-api/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// OmsServiceServer implements the gRPC server
type OmsUserServiceServer struct {
	pb.UnimplementedUserServiceServer
	DB *gorm.DB
}

var jwtSecret = []byte("uIqReXVOaBUb8hUCwMr4")

// Login generates a JWT token for authenticated users
func (s *OmsUserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Println("login------------------------")

	// Validate user credentials (replace with real validation)
	var user models.User
	err := s.DB.Where("email = ? AND password = ?", req.Email, req.Password).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %v", err) // Invalid credentials
	}

	// Create the JWT claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 1).Unix(), // Token expires in 1 hour
	}

	log.Println("claims--------------------------------", claims)

	// Generate the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	log.Println("Token:-", signedToken)
	if err != nil {
		return nil, fmt.Errorf("could not sign the token: %v", err)
	}

	// Validate the token right after creation (optional)
	tokenParsed, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return jwtSecret, nil // Return the secret for verification
	})

	// Check for token validity
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := tokenParsed.Claims.(jwt.MapClaims); ok && tokenParsed.Valid {
		log.Println("Token is valid. Claims:", claims)
	} else {
		return nil, fmt.Errorf("invalid token claims or token is not valid")
	}

	// Return the token in the response
	return &pb.LoginResponse{Token: signedToken}, nil
}

func (s *OmsUserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	// Initialize a new user from the request data
	newUser := models.User{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	// Insert the new user into the database using GORM
	if err := s.DB.Create(&newUser).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to insert user: %v", err)
	}

	// Return the newly created user details in the response using ToPb
	return newUser.ToPb(), nil
}

func (s *OmsUserServiceServer) GetUserById(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	// Find the user by ID in the database
	var user models.User
	if err := s.DB.Where("id = ? AND deleted_at IS NULL", req.GetUserId()).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Unable to fetch user data: %v", err)
	}

	// Return the user details in the response using ToPb
	return user.ToPb(), nil
}

func (s *OmsUserServiceServer) GetAllUsers(ctx context.Context, req *pb.EmptyRequestUser) (*pb.GetAllUsersResponse, error) {
	var users []models.User

	// Fetch non-deleted users from the database
	if err := s.DB.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to fetch users: %v", err)
	}

	// Map the users to gRPC response format
	var userResponses []*pb.User
	for _, user := range users {
		userResponses = append(userResponses, &pb.User{
			Id:        int32(user.ID),
			Name:      user.Name,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		})
	}

	// Return the users in the response
	return &pb.GetAllUsersResponse{
		Users: userResponses,
	}, nil
}

func (s *OmsUserServiceServer) UpdateUserById(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	// Find the user by ID
	var user models.User
	if err := s.DB.First(&user, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to fetch user: %v", err)
	}

	// Update user details
	user.Name = req.GetName()
	user.Email = req.GetEmail()

	// Save the updated user
	if err := s.DB.Save(&user).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update user: %v", err)
	}

	// Return the updated user in the response
	return &pb.User{
		Id:        int32(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	}, nil
}

func (s *OmsUserServiceServer) DeleteUserById(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	id := req.GetUserId()

	// Attempt to find the user, including soft-deleted users
	var user models.User
	if err := s.DB.Unscoped().First(&user, id).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "User not found: %v", err)
	}

	// Check if the user is already soft-deleted
	if user.DeletedAt.Valid {
		return &pb.DeleteUserResponse{Message: "User is already deleted"}, nil
	}

	// Proceed with soft delete (set deleted_at to the current time)
	user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	if err := s.DB.Save(&user).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete user: %v", err)
	}

	// Return success message in the response
	return &pb.DeleteUserResponse{Message: "User deleted successfully"}, nil
}

// GetUserOrders is the gRPC method for fetching user details and associated orders
func (s *OmsUserServiceServer) GetUserOrdersByUserId(ctx context.Context, req *pb.GetUserRequest) (*pb.UserOrderResponse, error) {
	id := req.GetUserId() // Retrieve user ID from the gRPC request

	// Fetch user details and associated orders in one query using Preload
	var user models.User
	if err := s.DB.Preload("Orders.Items").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Unable to fetch user data: %v", err)
	}

	// Check if the user is soft-deleted
	if user.DeletedAt.Valid {
		return nil, status.Errorf(codes.FailedPrecondition, "User has been soft deleted")
	}

	// Map orders and their items to response structs
	var ordersResponse []*pb.OrderResponseu
	for _, order := range user.Orders {
		orderResponse := &pb.OrderResponseu{
			Id:         int32(order.ID),
			TotalPrice: order.TotalPrice,
			FinalPrice: order.FinalPrice,
			Status:     order.Status,
		}

		// Map items to response struct
		var itemsResponse []*pb.ItemResponseu
		for _, item := range order.Items {
			itemsResponse = append(itemsResponse, &pb.ItemResponseu{
				ItemId:   item.ID,
				Price:    item.Price,
				Quantity: int32(item.Quantity),
			})
		}
		orderResponse.Items = itemsResponse
		ordersResponse = append(ordersResponse, orderResponse)
	}

	// Construct the full response combining user details and orders
	userResponse := &pb.UserOrderResponse{
		Id:            int32(user.ID),
		Name:          user.Name,
		Email:         user.Email,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     user.UpdatedAt.Format(time.RFC3339),
		DeletedAt:     user.DeletedAt.Time.Format(time.RFC3339),
		OrderResponse: ordersResponse,
	}

	// Send the final response with user and order details
	return userResponse, nil
}
