package app

import (
	"context"
	"log"

	"proxy-system-backend/internal/modules/filter"
	filterstore "proxy-system-backend/internal/storage/filter"
)

type FilterLoader struct {
	repo   filterstore.RuleRepository
	engine *filter.Engine
}

func NewFilterLoader(
	repo filterstore.RuleRepository,
	engine *filter.Engine,
) *FilterLoader {
	return &FilterLoader{repo: repo, engine: engine}
}

func (l *FilterLoader) Load(ctx context.Context) error {
	models, err := l.repo.List(ctx)
	if err != nil {
		return err
	}

	var compiled []*filter.CompiledRule

	for _, m := range models {
		rule, err := ModelToRule(m)
		if err != nil {
			log.Println("rule parse failed:", err)
			continue
		}

		cr, err := filter.CompileRule(*rule)
		if err != nil {
			log.Println("rule compile failed:", err)
			continue
		}

		compiled = append(compiled, cr)
	}

	l.engine.Replace(compiled)
	log.Printf("✅ filter rules loaded: %d\n", len(compiled))
	return nil
}
