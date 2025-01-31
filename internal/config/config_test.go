package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFonfigSuccess(t *testing.T) {
	t.Run("test config success", func(t *testing.T) {
		newAddress := "localhost:8080"
		args := []string{
			"-a=" + newAddress,
		}
		conf, err := NewConfig(args)
		require.NoError(t, err)
		require.Equal(t, newAddress, conf.Address)
	})
}

func TestConfigDefault(t *testing.T) {
	t.Run("test config default", func(t *testing.T) {
		args := []string{}
		conf, err := NewConfig(args)
		require.NoError(t, err)
		require.Equal(t, address, conf.Address)
	})
}
