package db

import (
	"context"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

	func Connect(dsn string) (*gorm.DB, error) {
		gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		sqlDB, err := gdb.DB()
		if err != nil {
			return nil, err
		}

		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			return nil, err
		}

		return gdb, nil
	}
