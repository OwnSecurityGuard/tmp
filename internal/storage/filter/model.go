package filterstore

import (
	"time"
)

type RuleModel struct {
	ID          int64 `gorm:"primaryKey"`
	Name        string
	Description string

	Action    string
	Direction uint8
	Priority  int
	Enabled   bool

	SrcCIDR string // JSON
	DstCIDR string
	SrcPort string
	DstPort string

	Tags string

	UpdatedAt time.Time
}
