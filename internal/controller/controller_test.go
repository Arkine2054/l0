package controller

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arkine2054/l0/internal/kafka"
	"github.com/Arkine2054/l0/internal/logic"
	"github.com/Arkine2054/l0/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestController_Index(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogic := NewMockLogic(ctrl)
	producer := &kafka.Producer{}

	var logicInterface logic.Logic = mockLogic
	controller := NewController(logicInterface, producer)

	router := gin.New()
	router.LoadHTMLGlob("../../templates/*")
	router.GET("/", controller.Index)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestController_GetOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		orderID        string
		mockSetup      func(m *MockLogic)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			orderID: "order123",
			mockSetup: func(m *MockLogic) {
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
					GetOrder(gomock.Any(), "order123").
					Return(order, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "empty order id",
			orderID: "",
			mockSetup: func(m *MockLogic) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "order not found",
			orderID: "missing",
			mockSetup: func(m *MockLogic) {
				m.EXPECT().
					GetOrder(gomock.Any(), "missing").
					Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "internal server error",
			orderID: "error",
			mockSetup: func(m *MockLogic) {
				m.EXPECT().
					GetOrder(gomock.Any(), "error").
					Return(nil, errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogic := NewMockLogic(ctrl)
			tt.mockSetup(mockLogic)

			producer := &kafka.Producer{}
			var logicInterface logic.Logic = mockLogic
			controller := NewController(logicInterface, producer)

			router := gin.New()
			router.GET("/order/*id", controller.GetOrder)

			w := httptest.NewRecorder()
			url := "/order/" + tt.orderID
			req, _ := http.NewRequest("GET", url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
