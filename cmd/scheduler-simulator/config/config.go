package config

import (
	"errors"
	"os"
	"strconv"

	"golang.org/x/xerrors"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config/env"
)

var (
	// ErrEmptyEnv represents the required environment variable don't exist.
	ErrEmptyEnv = errors.New("env is needed, but empty")
	// ErrInvalidEnv represents the environment variable is invalid value.
	ErrInvalidEnv = errors.New("invalid env")
)

// Config is configuration for simulator.
type Config struct {
	Env  env.Env
	Port int
}

// NewConfig gets some settings from environment variables.
func NewConfig() (*Config, error) {
	port, err := getPort()
	if err != nil {
		return nil, xerrors.Errorf("get port: %w", err)
	}

	e, err := getEnv()
	if err != nil {
		return nil, xerrors.Errorf("get env: %w", err)
	}

	return &Config{
		Env:  e,
		Port: port,
	}, nil
}

// getPort gets Port from the environment variable named PORT.
func getPort() (int, error) {
	p := os.Getenv("PORT")
	if p == "" {
		return 0, xerrors.Errorf("get PORT from env: %w", ErrEmptyEnv)
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, xerrors.Errorf("convert PORT of string to int: %w", err)
	}
	return port, nil
}

// getEnv gets Env from the environment variable named ENV.
func getEnv() (env.Env, error) {
	e := os.Getenv("ENV")
	if e == "" {
		return 0, xerrors.Errorf("get ENV from env: %w", ErrEmptyEnv)
	}

	switch e {
	case "development":
		return env.Development, nil
	case "production":
		return env.Production, nil
	}

	return 0, xerrors.Errorf("convert ENV of string to type Env: %w", ErrInvalidEnv)
}
