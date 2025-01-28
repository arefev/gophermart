package worker

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type request struct {
	accrualAddress string
}

func NewRequest(accrualAddress string) *request {
	return &request{accrualAddress: accrualAddress}
}

func (r *request) Request(ctx context.Context, number string, res *OrderResponse) error {
	client := resty.New()
	response, err := client.R().
		SetResult(res).
		SetContext(ctx).
		Get(r.getURL(number))

	if err != nil {
		return fmt.Errorf("request fail: %w", err)
	}

	res.Header = response.Header()
	res.HTTPStatus = response.StatusCode()

	return nil
}

func (r *request) getURL(number string) string {
	url := r.accrualAddress + "/api/orders/" + number
	if !strings.Contains(r.accrualAddress, "http://") {
		url = "http://" + url
	}

	return url
}
