package app

import (
	"encoding/json"
	"proxy-system-backend/internal/modules/filter"
	filterstore "proxy-system-backend/internal/storage/filter"
	"proxy-system-backend/internal/traffic"
)

func ModelToRule(m filterstore.RuleModel) (*filter.Rule, error) {
	var srcCIDR, dstCIDR []string
	var srcPort, dstPort []filter.PortRange
	var tags []string

	_ = json.Unmarshal([]byte(m.SrcCIDR), &srcCIDR)
	_ = json.Unmarshal([]byte(m.DstCIDR), &dstCIDR)
	_ = json.Unmarshal([]byte(m.SrcPort), &srcPort)
	_ = json.Unmarshal([]byte(m.DstPort), &dstPort)
	_ = json.Unmarshal([]byte(m.Tags), &tags)

	return &filter.Rule{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,

		Action:    filter.Action(m.Action),
		Direction: traffic.Direction(m.Direction),
		Priority:  m.Priority,
		Enabled:   m.Enabled,

		SrcCIDR: srcCIDR,
		DstCIDR: dstCIDR,
		SrcPort: srcPort,
		DstPort: dstPort,
		Tags:    tags,
	}, nil
}
