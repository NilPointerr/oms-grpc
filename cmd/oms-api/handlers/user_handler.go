package handlers

import (
	"context"
	"time"

	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	pb "github.com/keyurKalariya/OMS/cmd/oms-api/protobuf"
	"github.com/keyurKalariya/OMS/cmd/oms-api/utils"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// OmsServiceServer implements the gRPC server
type OmsUserServiceServer struct {
	pb.UnimplementedUserServiceServer
	DB *gorm.DB
}

// Register creates a new user account and returns a JWT token
func (s *OmsUserServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Validate input
	if req.GetName() == "" || req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name, email, and password are required")
	}

	// Check if user already exists
	var existingUser models.User
	if err := s.DB.Where("email = ?", req.GetEmail()).First(&existingUser).Error; err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	// Create new user
	newUser := models.User{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Password: string(hashedPassword),
	}

	// Insert the new user into the database
	if err := s.DB.Create(&newUser).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Generate JWT token
	token, err := utils.GenerateToken(newUser.ID, newUser.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	// Return response with token and user info
	return &pb.RegisterResponse{
		Token: token,
		User:  newUser.ToPb(),
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *OmsUserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate input
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}

	// Find user by email
	var user models.User
	if err := s.DB.Where("email = ? AND deleted_at IS NULL", req.GetEmail()).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch user: %v", err)
	}

	// Check if user has a password set (for existing users without passwords)
	if user.Password == "" {
		return nil, status.Errorf(codes.FailedPrecondition, "user account requires password reset. Please contact administrator or create a new account.")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword())); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid email or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	// Return response with token and user info
	return &pb.LoginResponse{
		Token: token,
		User:  user.ToPb(),
	}, nil
}

func (s *OmsUserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	// Validate input
	if req.GetName() == "" || req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name, email, and password are required")
	}

	// Check if user already exists
	var existingUser models.User
	if err := s.DB.Where("email = ?", req.GetEmail()).First(&existingUser).Error; err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	// Initialize a new user from the request data
	newUser := models.User{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Password: string(hashedPassword),
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
