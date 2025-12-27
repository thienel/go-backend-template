package database

import (
	"fmt"
	"sync"

	"github.com/thienel/tlog"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/thienel/go-backend-template/pkg/config"
)

var (
	db   *gorm.DB
	once sync.Once
)

// Init initializes the database connection
func Init(cfg *config.DatabaseConfig) error {
	var initErr error

	once.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.Password,
			cfg.DBName,
			cfg.SSLMode,
			cfg.TimeZone,
		)

		gormConfig := &gorm.Config{
			Logger: tlog.NewGormLogger(),
		}

		db, initErr = gorm.Open(postgres.Open(dsn), gormConfig)
		if initErr != nil {
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			initErr = err
			return
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		tlog.Info("Database connection established",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.DBName),
		)
	})

	return initErr
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// AutoMigrate runs auto migration
func AutoMigrate(models ...interface{}) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.AutoMigrate(models...)
}
