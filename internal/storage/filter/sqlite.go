package filterstore

import (
	"context"

	"gorm.io/gorm"
)

type SQLiteRepo struct {
	db *gorm.DB
}

func NewSQLiteRepo(db *gorm.DB) *SQLiteRepo {
	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) List(ctx context.Context) ([]RuleModel, error) {
	var rules []RuleModel
	err := r.db.WithContext(ctx).
		Order("priority desc").
		Find(&rules).Error
	return rules, err
}

func (r *SQLiteRepo) Save(ctx context.Context, m *RuleModel) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *SQLiteRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Delete(&RuleModel{}, id).Error
}
