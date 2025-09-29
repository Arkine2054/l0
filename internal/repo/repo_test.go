package repo

import (
	"context"
	"testing"
	"time"

	"github.com/Arkine2054/l0/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRepo_CreateOrder(t *testing.T) {
	// Этот тест требует реальной базы данных
	// В реальном проекте лучше использовать testcontainers или in-memory SQLite
	t.Skip("Skipping integration test - requires database")
}

func TestRepo_GetByID(t *testing.T) {
	// Этот тест требует реальной базы данных
	// В реальном проекте лучше использовать testcontainers или in-memory SQLite
	t.Skip("Skipping integration test - requires database")
}

func TestRepo_WarmUpCache(t *testing.T) {
	// Этот тест требует реальной базы данных
	// В реальном проекте лучше использовать testcontainers или in-memory SQLite
	t.Skip("Skipping integration test - requires database")
}

func TestRepo_Close(t *testing.T) {
	// Этот тест требует реальной базы данных
	// В реальном проекте лучше использовать testcontainers или in-memory SQLite
	t.Skip("Skipping integration test - requires database")
}

// Вспомогательные функции для создания тестовых данных
func createTestOrder() *models.Order {
	return &models.Order{
		OrderUID:        "test-order-123",
		TrackNumber:     "WBILMTESTTRACK",
		Entry:           "WBIL",
		Locale:          "en",
		CustomerID:      "test-customer-123",
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
}

// Unit test: cache hit path should return from cache without touching DB
func TestRepo_GetByID_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := NewMockCache(ctrl)
	want := &models.Order{OrderUID: "hit"}
	mockCache.EXPECT().Get("hit").Return(want, true)

	r := NewRepoWithCache(nil, mockCache)
	got, err := r.GetByID(context.Background(), "hit")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// Тесты для создания репозитория
func TestNewRepo(t *testing.T) {
	tests := []struct {
		name      string
		cacheSize int
		wantPanic bool
	}{
		{
			name:      "valid cache size",
			cacheSize: 100,
			wantPanic: false,
		},
		{
			name:      "zero cache size",
			cacheSize: 0,
			wantPanic: true,
		},
		{
			name:      "negative cache size",
			cacheSize: -1,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					NewRepo(nil, tt.cacheSize)
				})
			} else {
				assert.NotPanics(t, func() {
					NewRepo(nil, tt.cacheSize)
				})
			}
		})
	}
}
