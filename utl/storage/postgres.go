package storage

import (
	"context"

	pgDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TxKey string

const (
	GormTxKey TxKey = "GormTxKey"
)

// NewPostgresDB creates a new Postgres DB instance.
func NewPostgresDB() (*gorm.DB, error) {
	gdb, err := gorm.Open(pgDriver.Open("user=pgadmin password=pgadmin dbname=covax19 port=5432 sslmode=disable"), &gorm.Config{})

	return gdb, err
}

func GetGormDBFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if ctx.Value(GormTxKey) != nil {
		if val, ok := ctx.Value(GormTxKey).(*gorm.DB); ok {
			return val
		}
	}

	return db.WithContext(ctx)
}
