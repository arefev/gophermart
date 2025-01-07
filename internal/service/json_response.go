package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, data any) error {
	e := json.NewEncoder(w)
	if err := e.Encode(data); err != nil {
		return fmt.Errorf("json response ecode data fail: %w", err)
	}

	return nil
}