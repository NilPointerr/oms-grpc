package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	pb "github.com/keyurKalariya/OMS/cmd/oms-api/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type OrderServiceServer struct {
	pb.UnimplementedOrderServiceServer
	DB *gorm.DB
}

func (s *OrderServiceServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {

	// Bind the incoming request to the Order model
	var newOrder models.Order
	newOrder.UserID = req.GetOrder().UserId

	// Initialize total price and order items
	var totalPrice int32
	var orderItems []models.OrderItem

	// Loop through the OrderItems to calculate the total price
	for _, item := range req.GetOrder().Items {
		var itemRecord models.Item
		// Use GORM's First method to get the item by ID
		if err := s.DB.First(&itemRecord, item.ItemId).Error; err != nil {
			log.Println("Error fetching item for item ID", item.ItemId, ":", err)
			// Return error status with message
			return &pb.OrderResponse{
				OrderResponse: &pb.OrderResponse1{
					Status: "Failed to fetch item",
				},
			}, nil
		}

		// Extract the price from the fetched item record
		price := itemRecord.Price

		// Calculate the total price
		totalPrice += price * int32(item.Quantity)

		// Populate item details, including price
		orderItem := models.OrderItem{
			ItemID:   item.ItemId,
			Quantity: item.Quantity,
			Price:    float64(price),
		}
		// Add item to the orderItems array
		orderItems = append(orderItems, orderItem)
	}

	// Calculate discounts based on predefined conditions
	discounts := calculateDiscounts(s.DB, newOrder, orderItems)

	// Calculate the final price after applying discounts
	finalPrice := calculateTotalPrice(s.DB, orderItems, discounts)

	// Set the total and final price in the order object
	newOrder.TotalPrice = float64(totalPrice)
	newOrder.FinalPrice = finalPrice
	newOrder.Status = "Pending"

	// Insert the new order into the database
	if err := s.DB.Create(&newOrder).Error; err != nil {
		log.Println("Error inserting order:", err)
		// Return error status with message
		return &pb.OrderResponse{
			OrderResponse: &pb.OrderResponse1{
				Status: "Failed to insert order",
			},
		}, nil
	}

	// Insert items into the order_items table using GORM
	for _, item := range orderItems {
		item.OrderID = newOrder.ID
		if err := s.DB.Create(&item).Error; err != nil {
			log.Println("Error inserting order item:", err)
			// Return error status with message
			return &pb.OrderResponse{
				OrderResponse: &pb.OrderResponse1{
					Status: "Failed to insert order item",
				},
			}, nil
		}
	}

	// Insert user_id and order_id into the userOrder table
	if err := s.DB.Model(&models.UserOrder{}).Create(&models.UserOrder{
		UserID:  newOrder.UserID,
		OrderID: newOrder.ID,
	}).Error; err != nil {
		log.Println("Error inserting into userOrder table:", err)
		// Return error status with message
		return &pb.OrderResponse{
			OrderResponse: &pb.OrderResponse1{
				Status: "Failed to insert into userOrder table",
			},
		}, nil
	}

	// Create the response with order details
	var orderItemsResponse []*pb.OrderItemForResponse
	for _, item := range orderItems {
		orderItemsResponse = append(orderItemsResponse, &pb.OrderItemForResponse{
			ItemId:   item.ItemID,
			Quantity: item.Quantity,
			Price:    float64(item.Price),
		})
	}

	// Construct the OrderResponse1 struct first
	orderResponse := &pb.OrderResponse1{
		Id:         int32(newOrder.ID),
		UserId:     newOrder.UserID,
		TotalPrice: newOrder.TotalPrice,
		Status:     "Pending", // Status can be dynamic or based on conditions
		FinalPrice: finalPrice,
		Items:      orderItemsResponse,
	}

	// Return the response
	return &pb.OrderResponse{
		OrderResponse: orderResponse,
	}, nil
}

func (s *OrderServiceServer) GetAllOrders(ctx context.Context, req *pb.GetAllOrdersRequest) (*pb.AllOrderReponse, error) {
	var orders []models.Order

	// Fetch orders with GORM, excluding soft-deleted orders
	if err := s.DB.Where("deleted_at IS NULL").Find(&orders).Error; err != nil {
		log.Println("Error fetching orders:", err)
		return nil, status.Error(codes.Internal, "Unable to fetch orders")
	}

	var responseOrders []*pb.OrderResponse1

	// Iterate through each order to fetch associated order items
	for _, order := range orders {
		// Initialize the order response
		orderResponse := &pb.OrderResponse1{
			Id:         int32(order.ID),
			UserId:     int32(order.UserID),
			TotalPrice: order.TotalPrice,
			FinalPrice: order.FinalPrice,
			Status:     order.Status,
		}

		// Fetch order items using GORM
		var items []models.OrderItem
		if err := s.DB.Model(&models.OrderItem{}).Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			log.Println("Error fetching order items for order ID:", order.ID, err)
			return nil, status.Error(codes.Internal, "Unable to fetch order items")
		}

		// Create a map to aggregate items by ItemID
		itemMap := make(map[int32]*pb.OrderItemForResponse)

		// Iterate over order items and aggregate the data
		for _, item := range items {
			if existingItem, found := itemMap[int32(item.ItemID)]; found {
				// If the item already exists, update the quantity and price
				existingItem.Quantity += int32(item.Quantity)
				existingItem.Price += item.Price
			} else {
				// If the item does not exist, add it to the map
				itemMap[int32(item.ItemID)] = &pb.OrderItemForResponse{
					ItemId:   int32(item.ItemID),
					Quantity: int32(item.Quantity),
					Price:    item.Price,
				}
			}
		}

		// Convert the itemMap to a slice and add it to the response
		for _, aggregatedItem := range itemMap {
			orderResponse.Items = append(orderResponse.Items, aggregatedItem)
		}

		// Append the fully populated order response to the final response slice
		responseOrders = append(responseOrders, orderResponse)
	}

	// Return all orders with their aggregated items
	return &pb.AllOrderReponse{
		Orders: responseOrders,
	}, nil
}

func (s *OrderServiceServer) GetOrderById(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	// Fetch the order by ID, excluding soft-deleted records
	var order models.Order
	if err := s.DB.Where("id = ? AND deleted_at IS NULL", req.OrderId).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Order not found")
		}
		log.Println("Error fetching order:", err)
		return nil, status.Errorf(codes.Internal, "Unable to fetch order data")
	}

	// Fetch the order items for the specific order
	var items []models.OrderItem
	if err := s.DB.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		log.Println("Error fetching order items:", err)
		return nil, status.Errorf(codes.Internal, "Unable to fetch order items")
	}

	// Prepare the gRPC OrderResponse1 structure
	orderResponse := &pb.OrderResponse1{
		Id:         int32(order.ID),
		UserId:     int32(order.UserID),
		TotalPrice: order.TotalPrice,
		FinalPrice: order.FinalPrice,
		Status:     order.Status,
	}

	// Map the items to gRPC OrderItemForResponse
	for _, item := range items {
		orderResponse.Items = append(orderResponse.Items, &pb.OrderItemForResponse{
			ItemId:   int32(item.ItemID),
			Quantity: int32(item.Quantity),
			Price:    item.Price,
		})
	}

	// Return the response wrapped in OrderResponse
	return &pb.OrderResponse{
		OrderResponse: orderResponse,
	}, nil
}

func (s *OrderServiceServer) UpdateOrderById(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.OrderResponse1, error) {
	// Extract the order ID from the request
	orderID := req.GetOrderId()

	// Start a GORM transaction to ensure atomic updates
	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "Failed to start transaction")
	}
	defer tx.Rollback() // Rollback transaction in case of failure

	// Fetch the existing order
	var existingOrder models.Order
	if err := tx.Where("id = ? AND deleted_at IS NULL", orderID).First(&existingOrder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Order not found")
		}
		log.Println("Error fetching order:", err)
		return nil, status.Errorf(codes.Internal, "Failed to fetch order")
	}

	// Update the order status if it has changed
	if existingOrder.Status != "" && existingOrder.Status != "Confirmed" {
		if err := tx.Model(&existingOrder).Update("status", existingOrder.Status).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to update order status")
		}
	}

	// Delete all existing items for this order
	if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderItem{}).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete old order items")
	}

	// Process each item in the request
	var orderItemsForResponse []*pb.OrderItemForResponse
	for _, updatedItem := range req.GetItems() {
		var itemPrice float64

		// Fetch the price of the item from the database
		if err := tx.Model(&models.Item{}).Where("id = ?", updatedItem.GetItemId()).Pluck("price", &itemPrice).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.InvalidArgument, "Invalid item ID: %d", updatedItem.GetItemId())
			}
			return nil, status.Errorf(codes.Internal, "Failed to fetch item price")
		}

		// Insert the new item
		orderItem := models.OrderItem{
			OrderID:  orderID,
			ItemID:   updatedItem.GetItemId(),
			Quantity: updatedItem.GetQuantity(),
			Price:    itemPrice,
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to insert order item")
		}

		// Append the item to the response list
		orderItemsForResponse = append(orderItemsForResponse, &pb.OrderItemForResponse{
			ItemId:   updatedItem.GetItemId(),
			Quantity: updatedItem.GetQuantity(),
			Price:    itemPrice,
		})
	}

	// Recalculate the total price for the order
	var totalPrice float64
	if err := tx.Model(&models.OrderItem{}).Where("order_id = ?", orderID).Select("SUM(price * quantity)").Scan(&totalPrice).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to recalculate total price")
	}

	// Update the total price in the orders table
	if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Update("total_price", totalPrice).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update order total price")
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to commit transaction")
	}

	// Prepare and return the response
	response := &pb.OrderResponse1{
		Id:         int32(existingOrder.ID),
		UserId:     int32(existingOrder.UserID),
		TotalPrice: totalPrice,
		Status:     existingOrder.Status,
		FinalPrice: totalPrice, // Assuming no discounts are applied for simplicity
		Items:      orderItemsForResponse,
	}

	return response, nil
}

func (s *OrderServiceServer) UpdateOrderStatusByOrderId(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	// Extract the order ID from the request
	orderID := req.GetOrderId()

	// Start a GORM transaction to ensure atomic updates
	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "Failed to start transaction")
	}
	defer tx.Rollback() // Ensure rollback in case of failure

	// Fetch the current order
	var order models.Order
	if err := tx.Where("id = ? AND deleted_at IS NULL", orderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Order not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to fetch order")
	}

	// Check if the order status is 'Pending'
	if order.Status != "Pending" {
		return &pb.UpdateOrderStatusResponse{
			Message:       fmt.Sprintf("Order status is not 'Pending' (current status: %s)", order.Status),
			CurrentStatus: order.Status,
		}, nil
	}

	// Update the status to 'Confirm'
	if err := tx.Model(&order).Update("status", "Confirm").Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update order status")
	}

	// Commit the transaction if successful
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to commit transaction")
	}

	// Return the success response
	return &pb.UpdateOrderStatusResponse{
		Message:       "Order has been confirmed and placed successfully",
		CurrentStatus: "Confirm",
	}, nil
}

// DeleteOrderById deletes an order by its ID with soft delete functionality
func (s *OrderServiceServer) DeleteOrderById(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	id := req.GetOrderId()

	// Start a GORM transaction to ensure the order is deleted properly
	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", tx.Error)
	}
	defer tx.Rollback() // Ensure rollback in case of error

	// Fetch the existing order to check if it is already deleted
	var order models.Order
	if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.DeleteOrderResponse{
				Message: "Order not found",
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch order: %v", err)
	}

	// If the order is already deleted, return an error
	if order.DeletedAt.Valid {
		return &pb.DeleteOrderResponse{
			Message: "Order already deleted",
		}, nil
	}

	// Mark the order as deleted and update its status to "Cancelled"
	order.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	order.Status = "Cancelled"
	if err := tx.Save(&order).Error; err != nil {
		return nil, fmt.Errorf("failed to delete order: %v", err)
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Return the success response
	return &pb.DeleteOrderResponse{
		Message: "Order deleted and status set to 'Cancelled' successfully",
	}, nil
}

// UpdateOrderStatusByOrderId updates the order status to 'Confirm' if it is currently 'Pending'

func calculateDiscounts(db *gorm.DB, order models.Order, items []models.OrderItem) models.Discounts {
	discounts := models.Discounts{}

	// Seasonal discount (e.g., December 3 - December 31)
	currentDate := time.Now()
	if currentDate.Month() == time.December && currentDate.Day() >= 3 && currentDate.Day() <= 31 {
		discounts.SeasonalDiscount = 0.15
		log.Println("Seasonal discount applied: 15%")
	}

	// Volume-based discount (10 or more units of any single item)
	for _, item := range items {
		if item.Quantity >= 10 {
			volumeDiscount := 0.10 * item.Price * float64(item.Quantity)
			discounts.VolumeBasedDiscount += volumeDiscount
			log.Printf("Volume discount for item %d: %.2f", item.ItemID, volumeDiscount)
		}
	}

	// Loyalty discount (if the user has more than 5 orders)
	var orderCount int64
	// Use GORM to count the number of orders for the user
	err := db.Model(&models.Order{}).Where("user_id = ?", order.UserID).Count(&orderCount).Error
	if err != nil {
		log.Printf("Error fetching user order count: %v", err)
	}

	if orderCount >= 5 {
		discounts.LoyaltyDiscount = 0.05
		log.Println("Loyalty discount applied: 5%")
	}

	return discounts
}

func calculateTotalPrice(db *gorm.DB, items []models.OrderItem, discounts models.Discounts) float64 {
	var totalPrice float64
	for _, item := range items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	seasonalDiscount := totalPrice * discounts.SeasonalDiscount
	loyaltyDiscount := totalPrice * discounts.LoyaltyDiscount
	volumeDiscount := discounts.VolumeBasedDiscount

	totalDiscount := seasonalDiscount + loyaltyDiscount + volumeDiscount
	if totalDiscount > totalPrice {
		totalDiscount = totalPrice
	}

	finalPrice := totalPrice - totalDiscount
	log.Printf("Total Price: %.2f, Discounts: %.2f, Final Price: %.2f", totalPrice, totalDiscount, finalPrice)
	return finalPrice
}
