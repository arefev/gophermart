package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/logger"
	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/arefev/gophermart/internal/router"
	"github.com/arefev/gophermart/internal/service/worker"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	g, gCtx := errgroup.WithContext(mainCtx)

	zLog.Info("Worker starting...")
	g.Go(func() error {
		orderRep := repository.NewOrder(zLog)
		balanceRep := repository.NewBalance(zLog)
		worker.NewWorker(zLog, orderRep, balanceRep).Run(mainCtx)
		return nil
	})

	zLog.Info(
		"Server starting...",
		zap.String("address", conf.Address),
		zap.String("log level", conf.LogLevel),
	)

	server := http.Server{
		Addr:    conf.Address,
        Handler: router.New(zLog, &conf),
		BaseContext: func(_ net.Listener) context.Context {
			return mainCtx
		},
	}

	g.Go(func() error {
		return server.ListenAndServe()
	})

	g.Go(func() error {
		<-mainCtx.Done()
        return server.Shutdown(gCtx)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("exit reason %w", err)
	}

	return nil
}
