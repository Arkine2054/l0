package models

import "time"

type Order struct {
	OrderUID        string    `json:"order_uid" validate:"required"`
	TrackNumber     string    `json:"track_number" validate:"required"`
	Entry           string    `json:"entry" validate:"required"`
	Locale          string    `json:"locale" validate:"required"`
	CustomerID      string    `json:"customer_id" validate:"required"`
	DeliveryService string    `json:"delivery_service" validate:"required"`
	ShardKey        string    `json:"shardkey" validate:"required"`
	SmID            int       `json:"sm_id" validate:"gte=0"`
	DateCreated     time.Time `json:"date_created" validate:"required"`
	OofShard        string    `json:"oof_shard" validate:"required"`

	Delivery Delivery `json:"delivery" validate:"required"`
	Payment  Payment  `json:"payment" validate:"required"`
	Items    []Item   `json:"items" validate:"required,min=1,dive"`
}

type Delivery struct {
	Name    string `json:"name" validate:"required"`
	Phone   string `json:"phone" validate:"required"`
	Zip     string `json:"zip" validate:"required"`
	City    string `json:"city" validate:"required"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
}

type Payment struct {
	Transaction  string `json:"transaction" validate:"required"`
	Currency     string `json:"currency" validate:"required"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"required,gt=0"`
	PaymentDT    int64  `json:"payment_dt" validate:"required,gt=0"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int    `json:"goods_total" validate:"gte=0"`
	CustomFee    int    `json:"custom_fee" validate:"gte=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"required,gt=0"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"required,gt=0"`
	RID         string `json:"rid" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale" validate:"gte=0,lte=100"`
	Size        string `json:"size" validate:"required"`
	TotalPrice  int    `json:"total_price" validate:"required,gt=0"`
	NmID        int    `json:"nm_id" validate:"required,gt=0"`
	Brand       string `json:"brand" validate:"required"`
	Status      int    `json:"status" validate:"required,gte=0"`
}
