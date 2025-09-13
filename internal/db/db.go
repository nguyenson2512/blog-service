package db

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/example/blog-service/internal/config"
)

type Database struct {
	Gorm *gorm.DB
	SQL  *sql.DB
}

func Connect(cfg *config.Config) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode, cfg.DBTimezone)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	return &Database{Gorm: gormDB, SQL: sqlDB}, nil
}

func (d *Database) AutoMigrate(modelsToMigrate ...interface{}) error {
	return d.Gorm.AutoMigrate(modelsToMigrate...)
}

func (d *Database) EnsureGINIndexOnTags() error {
	return d.Gorm.Exec("CREATE INDEX IF NOT EXISTS idx_posts_tags_gin ON posts USING GIN (tags);").Error
}

func (d *Database) Close() error {
	if d.SQL != nil {
		return d.SQL.Close()
	}
	return nil
}

func (d *Database) Transaction(fc func(tx *gorm.DB) error) error {
	return d.Gorm.Transaction(fc)
} 