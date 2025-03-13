package postgres

import (
	"database/sql"
	"fmt"
	"strconv"

	"portal/internal/config"

	_ "github.com/lib/pq"
)

const (
	qrGetViews              = `SHOW portaldb.total_views;`
	qrTemplateIncreaseViews = `ALTER SYSTEM SET portaldb.total_views = `
	qrReloadConf            = `SELECT pg_reload_conf();`
)

type Storage struct {
	DB *sql.DB
}

func New(cfg config.SQL) (*Storage, error) {
	const op = "storage.postgres.New" // Имя текущей функции для логов и ошибок

	DB, err := sql.Open(cfg.PostgresDriver, cfg.PostgresInfo) // Подключаемся к БД
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: DB}, nil
}

func (s *Storage) IncreaseViews(amount int) error {
	const op = "storage.postgres.IncreaseViews"

	var views int
	err := s.DB.QueryRow(qrGetViews).Scan(&views)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	totalViews := views + amount
	qrIncreaseViews := qrTemplateIncreaseViews + strconv.Itoa(totalViews) + ";"
	_, err = s.DB.Exec(qrIncreaseViews)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.DB.Exec(qrReloadConf)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetViews() (int, error) {
	const op = "storage.postgres.IncreaseViews"

	var views int
	err := s.DB.QueryRow(qrGetViews).Scan(&views)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	return views, nil
}
