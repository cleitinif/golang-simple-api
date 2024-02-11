package config

import (
	"os"
	"participantes/cleitinif/errors"
	"strconv"
	"time"

	er "errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseConfig struct {
	Host            string
	Port            uint16
	User            string
	Password        string
	Database        string
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnTimeout     int64
	MinConnections  int32
	MaxConnections  int32
}

func NewDatabaseConfig() (*DatabaseConfig, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_DATABASE")
	minConnections := os.Getenv("DB_MIN_CONNECTIONS")
	maxConnections := os.Getenv("DB_MAX_CONNECTIONS")
	maxConnLifetime := os.Getenv("DB_MAX_CONN_LIFETIME")
	maxConnIdleTime := os.Getenv("DB_MAX_CONN_IDLE_TIME")

	if host == "" {
		return nil, er.New("DB_HOST variable is required")
	}

	if port == "" {
		return nil, er.New("DB_PORT variable is required")
	}
	convPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, er.New("DB_PORT must be a number")
	}

	if user == "" {
		return nil, er.New("DB_USER variable is required")
	}

	if password == "" {
		return nil, er.New("DB_PASSWORD variable is required")
	}

	if database == "" {
		return nil, er.New("DB_DATABASE variable is required")
	}

	if minConnections == "" {
		return nil, er.New("DB_MIN_CONNECTIONS variable is required")
	}
	convMinConnections, err := strconv.Atoi(minConnections)
	if err != nil {
		return nil, er.New("DB_MIN_CONNECTIONS must be a number")
	}

	if maxConnections == "" {
		return nil, er.New("DB_MAX_CONNECTIONS variable is required")
	}
	convMaxConnections, err := strconv.Atoi(maxConnections)
	if err != nil {
		return nil, er.New("DB_MAX_CONNECTIONS must be a number")
	}

	if maxConnLifetime == "" {
		return nil, er.New("DB_MAX_CONN_LIFETIME variable is required")
	}
	connMaxConnLifetime, err := strconv.Atoi(maxConnLifetime)
	if err != nil {
		return nil, er.New("DB_MAX_CONN_LIFETIME must be a number")
	}

	if maxConnIdleTime == "" {
		return nil, er.New("DB_MAX_CONN_IDLE_TIME variable is required")
	}
	connMaxConnIdleTime, err := strconv.Atoi(maxConnIdleTime)
	if err != nil {
		return nil, er.New("DB_MAX_CONN_IDLE_TIME must be a number")
	}

	return &DatabaseConfig{
		Host:            host,
		Port:            uint16(convPort),
		User:            user,
		Password:        password,
		Database:        database,
		MinConnections:  int32(convMinConnections),
		MaxConnections:  int32(convMaxConnections),
		MaxConnLifetime: time.Duration(connMaxConnLifetime),
		MaxConnIdleTime: time.Duration(connMaxConnIdleTime),
	}, nil
}

func HandleDatabaseError(err error) error {
	if err.Error() == "no rows in result set" {
		return errors.NewNotFoundError()
	}

	var pgErr *pgconn.PgError
	if er.As(err, &pgErr) && pgErr.Code == "40001" {
		return errors.NewTransactionConflictError()
	}

	return errors.NewInternalError()
}
