package filterstore

import "context"

type RuleRepository interface {
	List(ctx context.Context) ([]RuleModel, error)
	Save(ctx context.Context, r *RuleModel) error
	Delete(ctx context.Context, id int64) error
}
