package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name              string             `json:"name" bson:"name" validate:"required"`
	Description       string             `json:"description,omitempty" bson:"description"`
	Price             float64            `json:"price" bson:"price" validate:"required"`
	Rating            float64            `json:"rating,omitempty" bson:"rating"`
	Category          string             `json:"category,omitempty" bson:"category"`
	Disponibility     bool               `json:"disponibility" bson:"disponibility"`
	QuantityWarehouse int                `json:"quantity_warehouse" bson:"quantity_warehouse"`
	Images            []string           `json:"images" bson:"images"`
	Dimensions        Dimensions         `json:"dimensions" bson:"dimensions"`
	Brand             string             `json:"brand,omitempty" bson:"brand"`
	Colors            []string           `json:"colors" bson:"colors"`
	SKU               string             `json:"sku,omitempty" bson:"sku"`
	DateCreated       time.Time          `json:"dt_created" bson:"dt_created"`
	DateUpdated       time.Time          `json:"dt_updated" bson:"dt_updated"`
}

type Dimensions struct {
	Height float64 `json:"height,omitempty" bson:"height"`
	Width  float64 `json:"width,omitempty" bson:"width"`
	Length float64 `json:"length,omitempty" bson:"length"`
	Weight float64 `json:"weight,omitempty" bson:"weight"`
}

type CreateProductRequest struct {
	Name              string     `json:"name" validate:"required"`
	Description       string     `json:"description,omitempty"`
	Price             float64    `json:"price" validate:"required"`
	Rating            float64    `json:"rating,omitempty"`
	Category          string     `json:"category,omitempty"`
	Disponibility     bool       `json:"disponibility"`
	QuantityWarehouse int        `json:"quantity_warehouse"`
	Images            []string   `json:"images"`
	Dimensions        Dimensions `json:"dimensions"`
	Brand             string     `json:"brand,omitempty"`
	Colors            []string   `json:"colors"`
	SKU               string     `json:"sku,omitempty"`
}

type ProductResponse struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description,omitempty"`
	Price             float64    `json:"price"`
	Rating            float64    `json:"rating,omitempty"`
	Category          string     `json:"category,omitempty"`
	Disponibility     bool       `json:"disponibility"`
	QuantityWarehouse int        `json:"quantity_warehouse"`
	Images            []string   `json:"images"`
	Dimensions        Dimensions `json:"dimensions"`
	Brand             string     `json:"brand,omitempty"`
	Colors            []string   `json:"colors"`
	SKU               string     `json:"sku,omitempty"`
	DateCreated       time.Time  `json:"dt_created"`
	DateUpdated       time.Time  `json:"dt_updated"`
}

type CountResponse struct {
	Count int64 `json:"count"`
}

type PaginatedProductsResponse struct {
	Products   []ProductResponse `json:"products"`
	TotalCount int64             `json:"totalCount"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}
