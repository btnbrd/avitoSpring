package pg

import (
	"avitoSpring/internal/config"
	"avitoSpring/internal/storage"
	"database/sql"
	"fmt"
)

type Storage struct {
	DB             *sql.DB
	UserStorage    storage.UserStorage
	PVZStorage     storage.PVZStorageI
	ProductStorage storage.ProductStorage
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	userStorage, err := NewUserStorage(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create User storage: %w", err)
	}

	return &Storage{
		DB:             db,
		UserStorage:    userStorage,
		PVZStorage:     NewPVZStorage(db),
		ProductStorage: NewProductStorage(db),
	}, nil
}
