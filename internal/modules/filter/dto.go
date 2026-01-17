package filter

// internal/modules/filter/dto.go
type RuleDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Priority  int    `json:"priority"`
	Action    string `json:"action"` // allow | deny
	Direction string `json:"direction,omitempty"`

	SrcIP   string `json:"src_ip,omitempty"`
	DstIP   string `json:"dst_ip,omitempty"`
	SrcPort string `json:"src_port,omitempty"`
	DstPort string `json:"dst_port,omitempty"`

	Tags []string `json:"tags,omitempty"`

	Enabled   bool  `json:"enabled"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}
