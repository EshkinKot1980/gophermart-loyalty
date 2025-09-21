package dto

import (
	"testing"
)

func TestOrder_Validate(t *testing.T) {
	tests := []struct {
		name       string
		sentNumber string
		order      Order
		wantErr    bool
	}{
		{
			name:       "positive_status_processed",
			sentNumber: "5062821234567892",
			order: Order{
				Number:  "5062821234567892",
				Status:  OrderStatusProcessed,
				Accrual: 13.13,
			},
			wantErr: false,
		},
		{
			name:       "positive_status_invalid",
			sentNumber: "5062821234567892",
			order: Order{
				Number: "5062821234567892",
				Status: OrderStatusInvalid,
			},
			wantErr: false,
		},
		{
			name:       "positive_status_registred",
			sentNumber: "5062821234567892",
			order: Order{
				Number: "5062821234567892",
				Status: OrderStatusRegistred,
			},
			wantErr: false,
		},
		{
			name:       "positive_status_processing",
			sentNumber: "5062821234567892",
			order: Order{
				Number: "5062821234567892",
				Status: OrderStatusProcessing,
			},
			wantErr: false,
		},
		{
			name:       "negative_number",
			sentNumber: "5062821234567819",
			order: Order{
				Number:  "5062821234567892",
				Status:  OrderStatusProcessed,
				Accrual: 13.13,
			},
			wantErr: true,
		},
		{
			name:       "negative_status",
			sentNumber: "5062821234567819",
			order: Order{
				Number:  "5062821234567892",
				Status:  "UNSUPORTED",
				Accrual: 13.13,
			},
			wantErr: true,
		},
		{
			name:       "negative_accrual",
			sentNumber: "5062821234567819",
			order: Order{
				Number:  "5062821234567892",
				Status:  "UNSUPORTED",
				Accrual: -13.13,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.order.Validate(tt.sentNumber)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}
