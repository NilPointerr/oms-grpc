package handlers

import (
	"context"
	"time"

	pb "github.com/keyurKalariya/OMS/cmd/oms-api/protobuf"

	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// OmsServiceServer implements the gRPC server
type OmsItemServiceServer struct {
	pb.UnimplementedOmsItemServiceServer
	DB *gorm.DB
}

func (s *OmsItemServiceServer) CreateItem(ctx context.Context, req *pb.ItemRequest) (*pb.ItemResponse, error) {
	// Validate the fields (check for empty strings or invalid price)
	if req.Name == "" || req.Description == "" || req.Price <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "All fields must be filled and price must be positive")
	}

	// Create a new Item model instance from the request
	newItem := models.Item{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	// Use GORM to insert the new item into the database
	if err := s.DB.Create(&newItem).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to insert item: %v", err)
	}

	// Return the response with the new item details
	return newItem.ToPb(), nil
}
func (s *OmsItemServiceServer) GetItemById(ctx context.Context, req *pb.GetItemRequest) (*pb.ItemResponse, error) {
	// Validate the request (check if ID is provided)
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Item ID is required")
	}

	var item models.Item

	// Fetch the item by ID using GORM
	if err := s.DB.Where("id = ? AND deleted_at IS NULL", req.Id).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "Item not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to fetch item: %v", err)
	}

	// Convert the item to a gRPC response and return
	return item.ToPb(), nil
}

func (s *OmsItemServiceServer) GetAllItems(ctx context.Context, req *pb.EmptyRequest) (*pb.GetAllItemResponse, error) {
	var items []models.Item

	// Fetch all non-deleted items from the database
	if err := s.DB.Where("deleted_at IS NULL").Find(&items).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to fetch data: %v", err)
	}

	// Convert the list of items to gRPC responses
	itemResponses := make([]*pb.ItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = item.ToPb()
	}

	// Return the response with the list of items
	return &pb.GetAllItemResponse{Items: itemResponses}, nil
}

func (s *OmsItemServiceServer) UpdateItemById(ctx context.Context, req *pb.UpdateItemRequest) (*pb.ItemResponse, error) {
	// Find the item by ID from the database
	var item models.Item
	if err := s.DB.First(&item, req.GetId()).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "Item not found: %v", err)
	}

	// Update the item's fields based on the request
	item.Name = req.GetName()
	item.Description = req.GetDescription()
	item.Price = req.GetPrice()
	item.UpdatedAt = time.Now() // Ensure UpdatedAt is set to the current time

	// Save the updated item back to the database
	if err := s.DB.Save(&item).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update item: %v", err)
	}

	// Convert the updated item to a protobuf response
	itemResponse := item.ToPb()

	// Return the updated item as a response
	return itemResponse, nil
}

func (s *OmsItemServiceServer) DeleteItemById(ctx context.Context, req *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	var item models.Item

	// Attempt to find the item, including soft-deleted items
	if err := s.DB.Unscoped().First(&item, req.GetItemId()).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "Item not found: %v", err)
	}

	// Check if the item is already soft-deleted
	if item.DeletedAt.Valid {
		return &pb.DeleteItemResponse{Message: "Item is already deleted"}, nil
	}

	// Proceed with soft delete (setting deleted_at to the current time)
	item.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	if err := s.DB.Save(&item).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete item: %v", err)
	}

	// Return the success message in the response
	return &pb.DeleteItemResponse{Message: "Item deleted successfully"}, nil
}
