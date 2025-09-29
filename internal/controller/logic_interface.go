package controller

import (
	"context"

	"github.com/Arkine2054/l0/internal/models"
)

//go:generate mockgen -destination logic_mock.go -source logic_interface.go -package controller

type Logic interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrder(ctx context.Context, id string) (*models.Order, error)
}
