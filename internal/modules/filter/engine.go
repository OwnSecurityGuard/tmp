package filter

import (
	"proxy-system-backend/internal/traffic"
	"sort"
	"sync/atomic"
)

type Engine struct {
	enabled       atomic.Bool
	defaultAction atomic.Value // Action
	rules         atomic.Value // []*CompiledRule
}

func NewEngine() *Engine {
	e := &Engine{}
	e.enabled.Store(false)
	e.defaultAction.Store(ActionAllow)
	e.rules.Store([]*CompiledRule{})
	return e
}
func (e *Engine) Load(cfg Config, rules []Rule) error {
	var compiled []*CompiledRule

	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		cr, err := CompileRule(r)
		if err != nil {
			return err
		}
		compiled = append(compiled, cr)
	}

	sort.Slice(compiled, func(i, j int) bool {
		return compiled[i].Priority > compiled[j].Priority
	})

	e.enabled.Store(cfg.Enabled)
	e.defaultAction.Store(cfg.DefaultAction)
	e.rules.Store(compiled)

	return nil
}

func (e *Engine) Match(ctx *traffic.PacketContext) bool {
	if !e.enabled.Load() {
		return true
	}

	rules := e.rules.Load().([]*CompiledRule)
	for _, r := range rules {
		if r.Match(ctx) {
			return r.Action == ActionAllow
		}
	}

	return e.defaultAction.Load().(Action) == ActionAllow
}
func (e *Engine) Replace(rules []*CompiledRule) {
	e.rules.Store(rules)
}
