package worker

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arefev/gophermart/internal/model"
	"github.com/go-resty/resty/v2"
)

type request struct {
	accrualAddress string
}

func NewRequest(accrualAddress string) *request {
	return &request{accrualAddress: accrualAddress}
}

func (r *request) Request(ctx context.Context, number string, res *OrderResponse) (time.Duration, error) {
	const waitTime = time.Duration(60) * time.Second

	client := resty.New()
	response, err := client.R().
		SetResult(res).
		SetContext(ctx).
		Get(r.getURL(number))

	if err != nil {
		return 0, fmt.Errorf("request fail: %w", err)
	}

	wait := time.Duration(0) * time.Second
	if response.StatusCode() == http.StatusTooManyRequests {
		t := response.Header().Get("Retry-After")
		d, err := time.ParseDuration(t + "s")
		if err != nil {
			d = waitTime
		}
		wait = d
	}

	if response.StatusCode() != http.StatusOK {
		res.Status = model.OrderStatusInvalid.String()
	}

	return wait, nil
}

func (r *request) getURL(number string) string {
	url := r.accrualAddress + "/api/orders/" + number
	if !strings.Contains(r.accrualAddress, "http://") {
		url = "http://" + url
	}

	return url
}
