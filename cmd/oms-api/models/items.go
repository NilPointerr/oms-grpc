package models

import (
	"time"

	pb "github.com/keyurKalariya/OMS/cmd/oms-api/protobuf"
	"gorm.io/gorm"
)

// Item represents an item in the OMS system
type Item struct {
	ID          int32          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       int32          `json:"price"`
	CreatedAt   time.Time      `json:"created_at"` // Change to time.Time
	UpdatedAt   time.Time      `json:"updated_at"` // Change to time.Time
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
}







// TableName sets the schema and table name for the Item model
// func (ItemNew) TableName() string {
//     return "oms.item_new" // Specify the full table name with the schema
// }

// ToPb converts the Item model to the protobuf ItemResponse
func (item *Item) ToPb() *pb.ItemResponse {
	return &pb.ItemResponse{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
	}
}
