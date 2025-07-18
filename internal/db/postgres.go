package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type DB struct {
	Conn *sql.DB
}

func NewPostgres(host, port, user, password, dbname string) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка проверки соединения: %w", err)
	}
	return &DB{Conn: conn}, nil
}
