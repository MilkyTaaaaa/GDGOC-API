package models

import(
	"time"
	"github.com/lib/pq"
)

type Menu struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name" validate:"required,min=3,max=255"`
	Category    string         `gorm:"type:varchar(100);not null;index" json:"category" validate:"required,oneof=foods drinks desserts snacks"`
	Calories    *int           `gorm:"type:integer" json:"calories" validate:"omitempty,gte=0"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price" validate:"required,gt=0"`
	Ingredients pq.StringArray `gorm:"type:text[]" json:"ingredients" validate:"required,min=1"`
	Description string         `gorm:"type:text" json:"description" validate:"omitempty,max=1000"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Menu) TableName() string{
	return "menus"
}


type CreateMenuRequest struct {
	Name        string   `json:"name" validate:"required,min=3,max=255"`
	Category    string   `json:"category" validate:"required,oneof=foods drinks desserts snacks"`
	Calories    *int     `json:"calories" validate:"omitempty,gte=0"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	Ingredients []string `json:"ingredients" validate:"required,min=1"`
	Description string   `json:"description" validate:"omitempty,max=1000"`
}

type UpdateMenuRequest struct {
	Name        string   `json:"name" validate:"required,min=3,max=255"`
	Category    string   `json:"category" validate:"required,oneof=foods drinks desserts snacks"`
	Calories    *int     `json:"calories" validate:"omitempty,gte=0"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	Ingredients []string `json:"ingredients" validate:"required,min=1"`
	Description string   `json:"description" validate:"omitempty,max=1000"`
}

type MenuFilters struct {
	Query        string   `query:"q"`
	Category    string   `query:"category"`
	MinPrice    float64   `query:"min_price"`
	MaxPrice       float64  `query:"max_price"`
	MaxCalories int `query:"max_cal"`
	Page int   `query:"page"`
	PerPage	int	`query:"per_page"`
	Sort	string	`query:"sort"`
}

type PaginationMeta struct{
	Total	int64	`json:"total"`
	Page	int		`json:"page"`
	PerPage	int		`json:"per_page"`
	TotalPages	int	`json:"total_pages"`
}

type MenuListResponse struct{
	Data	[]Menu	`json:"data"`
	Pagination	*PaginationMeta	`json:"pagination,omitempty"`
}

type MenuResponse struct{
	Message	string	`json:"message,omitempty"`
	Data	Menu	`json:"data"`
}

type ErrorResponse struct{
	Message	string	`json:"message"`
	Errors	interface{}	`json:"errors,omitempty"`
}

type MessageResponse struct{
	Message	string	`json:"message"`
}

type GroupByCategoryResponse struct{
	Data	interface{}	`json:"data"`
}