package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func Build(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("zap logger parse level failed: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("zap logger build from config failed: %w", err)
	}

	return zl, nil
}
