package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderStatusFromString(t *testing.T) {
	type want struct {
		status OrderStatus
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "status NEW",
			want: want{
				status: OrderStatusNew,
			},
		},
		{
			name: "status INVALID",
			want: want{
				status: OrderStatusInvalid,
			},
		},
		{
			name: "status PROCESSED",
			want: want{
				status: OrderStatusProcessed,
			},
		},
		{
			name: "status PROCESSING",
			want: want{
				status: OrderStatusProcessing,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := OrderStatusFromString(tt.want.status.String())

			require.Equal(t, status, tt.want.status)
		})
	}
}
