package postgres

import (
	"context"
	"time"

	"atom-maintenance/internal/domain"

	"gorm.io/gorm"
)

type txKey struct{}

type TxManager struct{ db *gorm.DB }

func NewTxManager(db *gorm.DB) domain.TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

func dbFrom(ctx context.Context, base *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return base
}

func withTimeout(ctx context.Context, base *gorm.DB, ttl time.Duration) (*gorm.DB, context.Context, func()) {
	db := dbFrom(ctx, base)
	if db != base {
		return db, ctx, func() {}
	}
	dbCtx, cancel := context.WithTimeout(ctx, ttl)
	return db.WithContext(dbCtx), dbCtx, cancel
}
