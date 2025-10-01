package util

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Arkine2054/l0/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalOrder(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		want    *models.Order
	}{
		{
			name: "valid order",
			data: func() []byte {
				order := &models.Order{
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
				data, _ := json.Marshal(order)
				return data
			}(),
			wantErr: false,
			want: &models.Order{
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
			},
		},
		{
			name:    "invalid json",
			data:    []byte(`{"invalid": json}`),
			wantErr: true,
			want:    nil,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
			want:    nil,
		},
		{
			name:    "null data",
			data:    []byte("null"),
			wantErr: false,
			want:    &models.Order{},
		},
		{
			name:    "partial order",
			data:    []byte(`{"order_uid": "test-123"}`),
			wantErr: false,
			want: &models.Order{
				OrderUID: "test-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalOrder(tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				if tt.name == "valid order" {
					assert.Equal(t, tt.want.OrderUID, got.OrderUID)
					assert.Equal(t, tt.want.TrackNumber, got.TrackNumber)
					assert.Equal(t, tt.want.Entry, got.Entry)
					assert.Equal(t, tt.want.Locale, got.Locale)
					assert.Equal(t, tt.want.CustomerID, got.CustomerID)
					assert.Equal(t, tt.want.DeliveryService, got.DeliveryService)
					assert.Equal(t, tt.want.ShardKey, got.ShardKey)
					assert.Equal(t, tt.want.SmID, got.SmID)
					assert.Equal(t, tt.want.OofShard, got.OofShard)

					assert.Equal(t, tt.want.Delivery.Name, got.Delivery.Name)
					assert.Equal(t, tt.want.Delivery.Phone, got.Delivery.Phone)
					assert.Equal(t, tt.want.Delivery.Zip, got.Delivery.Zip)
					assert.Equal(t, tt.want.Delivery.City, got.Delivery.City)
					assert.Equal(t, tt.want.Delivery.Address, got.Delivery.Address)
					assert.Equal(t, tt.want.Delivery.Region, got.Delivery.Region)
					assert.Equal(t, tt.want.Delivery.Email, got.Delivery.Email)

					assert.Equal(t, tt.want.Payment.Transaction, got.Payment.Transaction)
					assert.Equal(t, tt.want.Payment.Currency, got.Payment.Currency)
					assert.Equal(t, tt.want.Payment.Provider, got.Payment.Provider)
					assert.Equal(t, tt.want.Payment.Amount, got.Payment.Amount)
					assert.Equal(t, tt.want.Payment.PaymentDT, got.Payment.PaymentDT)
					assert.Equal(t, tt.want.Payment.Bank, got.Payment.Bank)
					assert.Equal(t, tt.want.Payment.DeliveryCost, got.Payment.DeliveryCost)
					assert.Equal(t, tt.want.Payment.GoodsTotal, got.Payment.GoodsTotal)
					assert.Equal(t, tt.want.Payment.CustomFee, got.Payment.CustomFee)

					assert.Len(t, got.Items, len(tt.want.Items))
					for i, item := range tt.want.Items {
						assert.Equal(t, item.ChrtID, got.Items[i].ChrtID)
						assert.Equal(t, item.TrackNumber, got.Items[i].TrackNumber)
						assert.Equal(t, item.Price, got.Items[i].Price)
						assert.Equal(t, item.RID, got.Items[i].RID)
						assert.Equal(t, item.Name, got.Items[i].Name)
						assert.Equal(t, item.Sale, got.Items[i].Sale)
						assert.Equal(t, item.Size, got.Items[i].Size)
						assert.Equal(t, item.TotalPrice, got.Items[i].TotalPrice)
						assert.Equal(t, item.NmID, got.Items[i].NmID)
						assert.Equal(t, item.Brand, got.Items[i].Brand)
						assert.Equal(t, item.Status, got.Items[i].Status)
					}
				} else {
					assert.NotNil(t, got)
				}
			}
		})
	}
}
