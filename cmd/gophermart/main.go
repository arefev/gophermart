package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/logger"
	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/arefev/gophermart/internal/router"
	"github.com/arefev/gophermart/internal/service/worker"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	conf, err := config.NewConfig(os.Args[1:])
	if err != nil {
		return fmt.Errorf("run: init config fail: %w", err)
	}

	zLog, err := logger.Build(conf.LogLevel)
	if err != nil {
		return fmt.Errorf("run: init logger fail: %w", err)
	}

	if err := db.Connect(conf.DatabaseDSN, zLog); err != nil {
		return fmt.Errorf("run: db connect fail: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			zLog.Error("db close failed: %w", zap.Error(err))
		}
	}()

	zLog.Info(
		"Server starting...",
		zap.String("address", conf.Address),
		zap.String("log level", conf.LogLevel),
	)

	orderRep := repository.NewOrder(zLog)
	balanceRep := repository.NewBalance(zLog)
	go worker.NewWorker(zLog, orderRep, balanceRep).Run()

	return fmt.Errorf("run: server start fail: %w", http.ListenAndServe(conf.Address, router.New(zLog, &conf)))
}
