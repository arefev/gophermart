package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/config"
	"github.com/arefev/gophermart/internal/db/postgresql"
	"github.com/arefev/gophermart/internal/logger"
	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/router"
	"github.com/arefev/gophermart/internal/trm"
	"github.com/arefev/gophermart/internal/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// "github.com/jmoiron/sqlx"
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

	db, err := postgresql.NewDB(zLog).Connect(conf.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("run: db trm connect fail: %w", err)
	}

	err = migrationsUp(conf.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("run: migration up fail: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			zLog.Error("db close failed: %w", zap.Error(err))
		}
	}()

	tr := trm.NewTr(db.Connection())
	app := application.App{
		Rep: application.Repository{
			User:    repository.NewUser(tr, zLog),
			Order:   repository.NewOrder(tr, zLog),
			Balance: repository.NewBalance(tr, zLog),
		},
		TrManager: trm.NewTrm(tr, zLog),
		Log:       zLog,
		Conf:      &conf,
	}

	g, gCtx := errgroup.WithContext(mainCtx)

	zLog.Info("Worker starting...")
	g.Go(func() error {
		wRequest := worker.NewRequest(conf.AccrualAddress)
		return worker.NewWorker(&app, wRequest).Run(gCtx)
	})

	zLog.Info(
		"Server starting...",
		zap.String("address", conf.Address),
		zap.String("log level", conf.LogLevel),
	)

	server := http.Server{
		Addr:    conf.Address,
		Handler: router.New(&app),
		BaseContext: func(_ net.Listener) context.Context {
			return gCtx
		},
	}

	g.Go(server.ListenAndServe)

	g.Go(func() error {
		<-mainCtx.Done()
		zLog.Info("Server stopped")
		return server.Shutdown(gCtx)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("exit reason: %w", err)
	}

	return nil
}

func migrationsUp(dsn string) error {
	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		return fmt.Errorf("migrations instance fail: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations up fail: %w", err)
	}

	return nil
}
