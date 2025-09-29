package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestConnection(t *testing.T) {
	t.Skip("Skipping Kafka integration test - requires Kafka")
	tests := []struct {
		name    string
		brokers string
		wantErr bool
	}{
		{
			name:    "valid brokers",
			brokers: "localhost:9092",
			wantErr: false,
		},
		{
			name:    "empty brokers",
			brokers: "",
			wantErr: true,
		},
		{
			name:    "invalid brokers",
			brokers: "invalid:port",
			wantErr: true,
		},
		{
			name:    "multiple brokers",
			brokers: "localhost:9092,localhost:9093",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TestConnection(tt.brokers)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// В реальном тесте без Kafka это может упасть
				// В production лучше использовать testcontainers
				if err != nil {
					t.Skip("Skipping test - Kafka not available")
					return
				}
				assert.NoError(t, err)
			}
		})
	}
}
