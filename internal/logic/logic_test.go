package logic

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Arkine2054/l0/internal/models"
	"github.com/Arkine2054/l0/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLogic_GetOrder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		mockSetup func(m *repo.MockRepo)
		id        string
		wantErr   error
		wantOrder *models.Order
	}{
		{
			name: "success",
			mockSetup: func(m *repo.MockRepo) {
				m.EXPECT().
					GetByID(ctx, "order123").
					Return(&models.Order{OrderUID: "order123"}, nil)
			},
			id:        "order123",
			wantErr:   nil,
			wantOrder: &models.Order{OrderUID: "order123"},
		},
		{
			name: "not found",
			mockSetup: func(m *repo.MockRepo) {
				m.EXPECT().
					GetByID(ctx, "missing").
					Return(nil, sql.ErrNoRows)
			},
			id:        "missing",
			wantErr:   sql.ErrNoRows,
			wantOrder: nil,
		},
		{
			name: "db error",
			mockSetup: func(m *repo.MockRepo) {
				m.EXPECT().
					GetByID(ctx, "broken").
					Return(nil, errors.New("db connection failed"))
			},
			id:        "broken",
			wantErr:   errors.New("db connection failed"),
			wantOrder: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repo.NewMockRepo(ctrl)
			tt.mockSetup(mockRepo)

			l := NewLogic(mockRepo)

			order, err := l.GetOrder(ctx, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantOrder, order)
		})
	}
}

func TestLogic_CreateOrder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		mockSetup func(m *repo.MockRepo)
		order     *models.Order
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(m *repo.MockRepo) {
				order := &models.Order{
					OrderUID:        "order123",
					TrackNumber:     "track123",
					Entry:           "WBIL",
					Locale:          "en",
					CustomerID:      "customer123",
					DeliveryService: "meest",
					ShardKey:        "9",
					SmID:            99,
					DateCreated:     time.Now(),
					OofShard:        "1",
					Delivery: models.Delivery{
						Name:    "Test Testov",
						Phone:   "+9720000000",
						Zip:     "2639809",
						City:    "Kiryat Mozkin",
						Address: "Ploshad Mira 15",
						Region:  "Kraiot",
						Email:   "test@gmail.com",
					},
					Payment: models.Payment{
						Transaction:  "b563feb7b2b84b6test",
						Currency:     "USD",
						Provider:     "wbpay",
						Amount:       1817,
						PaymentDT:    1637907727,
						Bank:         "alpha",
						DeliveryCost: 1500,
						GoodsTotal:   317,
						CustomFee:    0,
					},
					Items: []models.Item{
						{
							ChrtID:      9934930,
							TrackNumber: "WBILMTESTTRACK",
							Price:       453,
							RID:         "ab4219087a764ae0b229c92aa27dd3ff",
							Name:        "Mascaras",
							Sale:        30,
							Size:        "0",
							TotalPrice:  317,
							NmID:        2389212,
							Brand:       "Vivienne Sabo",
							Status:      202,
						},
					},
				}
				m.EXPECT().
					CreateOrder(ctx, order).
					Return(nil)
			},
			order: &models.Order{
				OrderUID:        "order123",
				TrackNumber:     "track123",
				Entry:           "WBIL",
				Locale:          "en",
				CustomerID:      "customer123",
				DeliveryService: "meest",
				ShardKey:        "9",
				SmID:            99,
				DateCreated:     time.Now(),
				OofShard:        "1",
				Delivery: models.Delivery{
					Name:    "Test Testov",
					Phone:   "+9720000000",
					Zip:     "2639809",
					City:    "Kiryat Mozkin",
					Address: "Ploshad Mira 15",
					Region:  "Kraiot",
					Email:   "test@gmail.com",
				},
				Payment: models.Payment{
					Transaction:  "b563feb7b2b84b6test",
					Currency:     "USD",
					Provider:     "wbpay",
					Amount:       1817,
					PaymentDT:    1637907727,
					Bank:         "alpha",
					DeliveryCost: 1500,
					GoodsTotal:   317,
					CustomFee:    0,
				},
				Items: []models.Item{
					{
						ChrtID:      9934930,
						TrackNumber: "WBILMTESTTRACK",
						Price:       453,
						RID:         "ab4219087a764ae0b229c92aa27dd3ff",
						Name:        "Mascaras",
						Sale:        30,
						Size:        "0",
						TotalPrice:  317,
						NmID:        2389212,
						Brand:       "Vivienne Sabo",
						Status:      202,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			mockSetup: func(m *repo.MockRepo) {
				order := &models.Order{
					OrderUID: "order123",
				}
				m.EXPECT().
					CreateOrder(ctx, order).
					Return(errors.New("database error"))
			},
			order: &models.Order{
				OrderUID: "order123",
			},
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repo.NewMockRepo(ctrl)
			tt.mockSetup(mockRepo)

			l := NewLogic(mockRepo)

			err := l.CreateOrder(ctx, tt.order)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
