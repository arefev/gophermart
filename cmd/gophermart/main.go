package main

import (
	"fmt"
	"log"
	"os"

	"github.com/arefev/gophermart/internal/logger"
)

func main() {
	if err := run(); err!= nil {
		log.Fatal(err)
	}
}

func run() error {
	config, err := NewConfig(os.Args[0:1])
	if err != nil {
        return fmt.Errorf("run: init config fail: %w", err)
    }

	zLog, err := logger.Build(config.LogLevel)
	if err != nil {
        return fmt.Errorf("run: init logger fail: %w", err)
    }

	zLog.Sugar().Infof("Server started on %s", config.Address)
	return nil
}
