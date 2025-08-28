package logic

import (
	"context"
	"github.com/Arkine2054/l0/internal/models"
	"github.com/Arkine2054/l0/internal/repo"
)

type logic struct {
	repo repo.Repo
}

type Logic interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrder(ctx context.Context, id string) (*models.Order, error)
}

func NewLogic(repo repo.Repo) Logic {
	return &logic{repo: repo}
}

func (l *logic) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	return l.repo.GetByID(ctx, id)
}

func (l *logic) CreateOrder(ctx context.Context, order *models.Order) error {
	return l.repo.CreateOrder(ctx, order)
}
