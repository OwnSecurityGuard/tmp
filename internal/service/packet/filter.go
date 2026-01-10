package packet

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"proxy-system-backend/internal/types"
)

// PacketFilter 数据包过滤器
type PacketFilter struct {
	config     atomic.Value // 存储 FilterConfig
	stats      atomic.Value // 存储 FilterStats
	rules      []*FilterRuleInternal
	mu         sync.RWMutex
	lastUpdate time.Time
}

// FilterRuleInternal 内部过滤规则结构
type FilterRuleInternal struct {
	ID            int
	Name          string
	Action        types.FilterAction
	Direction     types.FilterDirection
	SourceIPs     []net.IP
	SourceIPMasks []int
	SourcePorts   []PortRange
	DestIPs       []net.IP
	DestIPMasks   []int
	DestPorts     []PortRange
	Enabled       bool
	Priority      int
	Description   string
}

// PortRange 端口范围结构
type PortRange struct {
	Min int
	Max int
}

// NewPacketFilter 创建新的数据包过滤器
func NewPacketFilter() *PacketFilter {
	pf := &PacketFilter{
		lastUpdate: time.Now(),
	}

	// 初始化默认配置
	defaultConfig := types.FilterConfig{
		Enabled:       false,
		DefaultAction: types.FilterActionAllow,
		Rules:         []types.FilterRule{},
		UpdateTime:    time.Now(),
	}
	pf.config.Store(defaultConfig)

	// 初始化默认统计
	defaultStats := types.FilterStats{
		TotalPackets:   0,
		AllowedPackets: 0,
		BlockedPackets: 0,
		ActiveRules:    0,
		LastUpdateTime: time.Now(),
	}
	pf.stats.Store(defaultStats)

	return pf
}

// LoadConfig 加载过滤配置
func (pf *PacketFilter) LoadConfig(config types.FilterConfig) error {
	if err := pf.validateConfig(config); err != nil {
		return err
	}

	compiled, err := pf.compileRules(config.Rules)
	if err != nil {
		return err
	}

	pf.mu.Lock()
	defer pf.mu.Unlock()

	config.UpdateTime = time.Now()
	pf.config.Store(config)

	pf.rules = compiled
	pf.lastUpdate = time.Now()

	stats := pf.stats.Load().(types.FilterStats)
	stats.ActiveRules = len(compiled)
	stats.LastUpdateTime = time.Now()
	pf.stats.Store(stats)

	log.Printf("✅ Filter loaded: %d active rules", len(compiled))
	return nil
}

// compileRules 编译规则为内部格式
func (pf *PacketFilter) compileRules(rules []types.FilterRule) ([]*FilterRuleInternal, error) {
	var compiled []*FilterRuleInternal

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		compiledRule, err := pf.compileRule(rule)
		if err != nil {
			log.Printf("Warning: failed to compile rule %s: %v", rule.Name, err)
			continue
		}

		compiled = append(compiled, compiledRule)
	}

	// 按优先级排序
	sort.Slice(compiled, func(i, j int) bool {
		return compiled[i].Priority > compiled[j].Priority
	})

	return compiled, nil
}

// compileRule 编译单个规则
func (pf *PacketFilter) compileRule(rule types.FilterRule) (*FilterRuleInternal, error) {
	compiled := &FilterRuleInternal{
		ID:          rule.ID,
		Name:        rule.Name,
		Action:      rule.Action,
		Direction:   rule.Direction,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		Description: rule.Description,
	}

	// 编译源IP
	for _, ipAddr := range rule.SourceIPs {
		ip := net.ParseIP(ipAddr.IP)
		if ip == nil {
			return nil, fmt.Errorf("invalid source IP: %s", ipAddr.IP)
		}
		compiled.SourceIPs = append(compiled.SourceIPs, ip)
		compiled.SourceIPMasks = append(compiled.SourceIPMasks, ipAddr.Mask)
	}

	// 编译源端口
	for _, portRange := range rule.SourcePorts {
		if portRange.Min < 0 || portRange.Max < 0 || portRange.Min > portRange.Max {
			return nil, fmt.Errorf("invalid source port range: %d-%d", portRange.Min, portRange.Max)
		}
		compiled.SourcePorts = append(compiled.SourcePorts, PortRange{Min: portRange.Min, Max: portRange.Max})
	}

	// 编译目标IP
	for _, ipAddr := range rule.DestIPs {
		ip := net.ParseIP(ipAddr.IP)
		if ip == nil {
			return nil, fmt.Errorf("invalid dest IP: %s", ipAddr.IP)
		}
		compiled.DestIPs = append(compiled.DestIPs, ip)
		compiled.DestIPMasks = append(compiled.DestIPMasks, ipAddr.Mask)
	}

	// 编译目标端口
	for _, portRange := range rule.DestPorts {
		if portRange.Min < 0 || portRange.Max < 0 || portRange.Min > portRange.Max {
			return nil, fmt.Errorf("invalid dest port range: %d-%d", portRange.Min, portRange.Max)
		}
		compiled.DestPorts = append(compiled.DestPorts, PortRange{Min: portRange.Min, Max: portRange.Max})
	}

	return compiled, nil
}

// FilterPacket 过滤数据包
func (pf *PacketFilter) FilterPacket(clientID string, remoteAddr string, data []byte, direction types.FilterDirection) bool {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	config := pf.config.Load().(types.FilterConfig)

	if !config.Enabled {
		return true // 过滤未启用，允许所有数据包
	}

	stats := pf.stats.Load().(types.FilterStats)
	stats.TotalPackets++
	pf.stats.Store(stats)

	// 检查每个规则
	for _, rule := range pf.rules {
		if !pf.ruleMatches(rule, remoteAddr, direction) {
			continue
		}

		// 规则匹配，根据动作决定
		pf.updateStats(rule.Action == types.FilterActionAllow)
		return rule.Action == types.FilterActionAllow
	}

	// 没有匹配到任何规则，使用默认动作
	pf.updateStats(config.DefaultAction == types.FilterActionAllow)
	return config.DefaultAction == types.FilterActionAllow
}

// ruleMatches 检查规则是否匹配
func (pf *PacketFilter) ruleMatches(rule *FilterRuleInternal, remoteAddr string, direction types.FilterDirection) bool {
	// 检查方向
	if !pf.directionMatches(rule.Direction, direction) {
		log.Printf("Debug ruleMatches: direction mismatch (rule: %s, packet: %s)", rule.Direction, direction)
		return false
	}

	// 解析远程地址
	host, portStr, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr // 如果没有端口，使用整个字符串
		portStr = "0"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Debug ruleMatches: failed to parse port: %v", err)
		return false
	}

	log.Printf("Debug ruleMatches: checking remoteAddr %s (host: %s, port: %d)", remoteAddr, host, port)

	// 检查源IP（对于入站连接，远程地址就是源地址）
	remoteIP := net.ParseIP(host)
	if remoteIP != nil && len(rule.SourceIPs) > 0 {
		log.Printf("Debug ruleMatches: checking source IPs (count: %d)", len(rule.SourceIPs))
		if !pf.ipMatches(remoteIP, rule.SourceIPs, rule.SourceIPMasks) {
			log.Printf("Debug ruleMatches: source IP did not match")
			return false
		}
	}

	// 对于入站连接，远程地址的端口就是源端口
	// 规则中指定的SourcePorts应该匹配远程地址的端口
	if len(rule.SourcePorts) > 0 {
		log.Printf("Debug ruleMatches: checking source ports (count: %d)", len(rule.SourcePorts))
		if !pf.portMatches(port, rule.SourcePorts) {
			log.Printf("Debug ruleMatches: source port %d did not match", port)
			return false
		}
	}

	// 检查目标IP（对于入站连接，本地地址就是目标地址）
	// 这里暂时使用简化逻辑，实际应该获取本地地址
	if len(rule.DestIPs) > 0 {
		log.Printf("Debug ruleMatches: skipping dest IP check (not implemented)")
	}

	// 对于入站连接，远程地址的端口就是源端口
	// 目标端口的检查在当前实现中被跳过，因为我们无法获取到真正的目标端口

	log.Printf("Debug ruleMatches: rule matched!")
	return true
}

// directionMatches 检查方向是否匹配
func (pf *PacketFilter) directionMatches(ruleDir types.FilterDirection, pktDir types.FilterDirection) bool {
	return ruleDir == pktDir || ruleDir == types.FilterDirectionBoth
}

// ipMatches 检查IP是否匹配
func (pf *PacketFilter) ipMatches(ip net.IP, ips []net.IP, masks []int) bool {
	for i, ruleIP := range ips {
		mask := 32
		if i < len(masks) {
			mask = masks[i]
		}

		log.Printf("Debug ipMatches: checking IP %s against rule IP %s with mask %d", ip.String(), ruleIP.String(), mask)

		if mask == 32 {
			// 精确匹配
			if ip.Equal(ruleIP) {
				log.Printf("Debug ipMatches: IP %s exactly matches rule IP %s", ip.String(), ruleIP.String())
				return true
			}
		} else {
			// CIDR匹配
			_, subnet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ruleIP.String(), mask))
			if err != nil {
				log.Printf("Debug ipMatches: failed to parse CIDR: %v", err)
				continue
			}
			if subnet.Contains(ip) {
				log.Printf("Debug ipMatches: IP %s is within subnet %s", ip.String(), subnet.String())
				return true
			}
		}
	}
	log.Printf("Debug ipMatches: IP %s did not match any rules", ip.String())
	return false
}

// portMatches 检查端口是否匹配
func (pf *PacketFilter) portMatches(port int, ranges []PortRange) bool {
	for _, r := range ranges {
		if port >= r.Min && port <= r.Max {
			return true
		}
	}
	return false
}

// updateStats 更新统计信息
func (pf *PacketFilter) updateStats(allowed bool) {
	stats := pf.stats.Load().(types.FilterStats)
	if allowed {
		stats.AllowedPackets++
	} else {
		stats.BlockedPackets++
	}
	pf.stats.Store(stats)
}

// GetStats 获取统计信息
func (pf *PacketFilter) GetStats() types.FilterStats {
	return pf.stats.Load().(types.FilterStats)
}

// GetConfig 获取当前配置
func (pf *PacketFilter) GetConfig() types.FilterConfig {
	return pf.config.Load().(types.FilterConfig)
}

// validateRule 验证单个规则
func (pf *PacketFilter) validateRule(rule types.FilterRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}

	if rule.Action != types.FilterActionAllow && rule.Action != types.FilterActionDeny {
		return fmt.Errorf("invalid action: %s", rule.Action)
	}

	if rule.Direction != types.FilterDirectionIn && rule.Direction != types.FilterDirectionOut && rule.Direction != types.FilterDirectionBoth {
		return fmt.Errorf("invalid direction: %s", rule.Direction)
	}

	// 验证IP格式
	for _, ipAddr := range rule.SourceIPs {
		if ip := net.ParseIP(ipAddr.IP); ip == nil {
			return fmt.Errorf("invalid source IP: %s", ipAddr.IP)
		}
		if ipAddr.Mask < 0 || ipAddr.Mask > 32 {
			return fmt.Errorf("invalid source IP mask: %d", ipAddr.Mask)
		}
	}

	for _, ipAddr := range rule.DestIPs {
		if ip := net.ParseIP(ipAddr.IP); ip == nil {
			return fmt.Errorf("invalid dest IP: %s", ipAddr.IP)
		}
		if ipAddr.Mask < 0 || ipAddr.Mask > 32 {
			return fmt.Errorf("invalid dest IP mask: %d", ipAddr.Mask)
		}
	}

	// 验证端口范围
	for _, portRange := range rule.SourcePorts {
		if portRange.Min < 0 || portRange.Max < 0 || portRange.Min > portRange.Max {
			return fmt.Errorf("invalid source port range: %d-%d", portRange.Min, portRange.Max)
		}
		if portRange.Max > 65535 {
			return fmt.Errorf("source port range exceeds maximum: %d", portRange.Max)
		}
	}

	for _, portRange := range rule.DestPorts {
		if portRange.Min < 0 || portRange.Max < 0 || portRange.Min > portRange.Max {
			return fmt.Errorf("invalid dest port range: %d-%d", portRange.Min, portRange.Max)
		}
		if portRange.Max > 65535 {
			return fmt.Errorf("dest port range exceeds maximum: %d", portRange.Max)
		}
	}

	return nil
}

// validateConfig 验证配置
func (pf *PacketFilter) validateConfig(config types.FilterConfig) error {
	if config.DefaultAction != types.FilterActionAllow && config.DefaultAction != types.FilterActionDeny {
		return fmt.Errorf("invalid default action: %s", config.DefaultAction)
	}

	for _, rule := range config.Rules {
		if err := pf.validateRule(rule); err != nil {
			return fmt.Errorf("invalid rule %s: %w", rule.Name, err)
		}
	}

	return nil
}

// AddRule 添加规则
func (pf *PacketFilter) AddRule(rule types.FilterRule) error {
	if err := pf.validateRule(rule); err != nil {
		return err
	}

	config := pf.GetConfig()
	config.Rules = append(config.Rules, rule)
	return pf.LoadConfig(config)
}

// UpdateRule 更新规则
func (pf *PacketFilter) UpdateRule(rule types.FilterRule) error {
	if err := pf.validateRule(rule); err != nil {
		return err
	}

	config := pf.GetConfig()
	for i, r := range config.Rules {
		if r.ID == rule.ID {
			config.Rules[i] = rule
			break
		}
	}
	return pf.LoadConfig(config)
}

// DeleteRule 删除规则
func (pf *PacketFilter) DeleteRule(ruleID int) error {
	config := pf.GetConfig()
	for i, r := range config.Rules {
		if r.ID == ruleID {
			config.Rules = append(config.Rules[:i], config.Rules[i+1:]...)
			break
		}
	}
	return pf.LoadConfig(config)
}

// EnableFiltering 启用/禁用过滤
func (pf *PacketFilter) EnableFiltering(enabled bool) {
	config := pf.GetConfig()
	config.Enabled = enabled
	pf.LoadConfig(config)
}

// SetDefaultAction 设置默认动作
func (pf *PacketFilter) SetDefaultAction(action types.FilterAction) {
	config := pf.GetConfig()
	config.DefaultAction = action
	pf.LoadConfig(config)
}
