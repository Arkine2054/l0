package util

import (
	"encoding/json"
	"github.com/Arkine2054/l0/internal/models"
)

func UnmarshalOrder(data []byte) (*models.Order, error) {
	var o models.Order
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	return &o, nil
}
