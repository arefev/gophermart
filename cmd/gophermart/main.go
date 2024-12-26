package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/arefev/gophermart/internal/logger"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/arefev/gophermart/internal/router"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config, err := NewConfig(os.Args[1:])
	if err != nil {
		return fmt.Errorf("run: init config fail: %w", err)
	}

	zLog, err := logger.Build(config.LogLevel)
	if err != nil {
		return fmt.Errorf("run: init logger fail: %w", err)
	}

	if err := db.Connect(config.DatabaseDSN); err != nil {
		return fmt.Errorf("run: db connect fail: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			zLog.Error("db close failed: %w", zap.Error(err))
		}
	}()

	zLog.Info(
		"Server starting...",
		zap.String("address", config.Address),
		zap.String("log level", config.LogLevel),
	)

	return fmt.Errorf("run: server start fail: %w", http.ListenAndServe(config.Address, router.New(zLog)))
}
