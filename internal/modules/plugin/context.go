package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// PluginContext 插件调用上下文
type PluginContext struct {
	// 上下文ID
	ID string `json:"id"`

	// 插件名称
	PluginName string `json:"plugin_name"`

	// 调用时间
	Timestamp time.Time `json:"timestamp"`

	// 自定义参数
	Params map[string]interface{} `json:"params"`

	// 超时时间
	Timeout time.Duration `json:"timeout"`

	// 是否启用详细日志
	Verbose bool `json:"verbose"`

	// 是否启用重试
	EnableRetry bool `json:"enable_retry"`

	// 最大重试次数
	MaxRetries int `json:"max_retries"`

	// 重试延迟
	RetryDelay time.Duration `json:"retry_delay"`

	// Go context 用于取消操作
	Context context.Context `json:"-"`
}

// NewPluginContext 创建新的插件调用上下文
func NewPluginContext(pluginName string) *PluginContext {
	return &PluginContext{
		ID:          generateContextID(),
		PluginName:  pluginName,
		Timestamp:   time.Now(),
		Params:      make(map[string]interface{}),
		Timeout:     time.Duration(GetTrafficHookTimeout()) * time.Millisecond,
		Verbose:     ShouldLogDecodeErrors(),
		EnableRetry: false,
		MaxRetries:  3,
		RetryDelay:  100 * time.Millisecond,
		Context:     context.Background(),
	}
}

// SetParam 设置参数
func (pc *PluginContext) SetParam(key string, value interface{}) *PluginContext {
	pc.Params[key] = value
	return pc
}

// GetParam 获取参数
func (pc *PluginContext) GetParam(key string) (interface{}, bool) {
	val, ok := pc.Params[key]
	return val, ok
}

// SetTimeout 设置超时时间
func (pc *PluginContext) SetTimeout(timeout time.Duration) *PluginContext {
	pc.Timeout = timeout
	return pc
}

// SetVerbose 设置详细日志
func (pc *PluginContext) SetVerbose(verbose bool) *PluginContext {
	pc.Verbose = verbose
	return pc
}

// EnableRetryWithConfig 启用重试
func (pc *PluginContext) EnableRetryWithConfig(maxRetries int, delay time.Duration) *PluginContext {
	pc.EnableRetry = true
	pc.MaxRetries = maxRetries
	pc.RetryDelay = delay
	return pc
}

// SetContext 设置 Go context
func (pc *PluginContext) SetContext(ctx context.Context) *PluginContext {
	pc.Context = ctx
	return pc
}

// ToJSON 转换为 JSON
func (pc *PluginContext) ToJSON() (string, error) {
	data, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal plugin context: %w", err)
	}
	return string(data), nil
}

// String 返回字符串表示
func (pc *PluginContext) String() string {
	return fmt.Sprintf("PluginContext{id=%s, plugin=%s, params=%d}",
		pc.ID, pc.PluginName, len(pc.Params))
}

// generateContextID 生成唯一的上下文ID
func generateContextID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// DecodeCallContext 解码调用上下文（不与 protobuf 的 DecodeRequest 冲突）
type DecodeCallContext struct {
	// 插件上下文
	Context *PluginContext `json:"context"`

	// 是否为客户端数据
	IsClient bool `json:"is_client"`

	// 待解码的数据
	Payload []byte `json:"payload"`

	// 额外的元数据
	Metadata map[string]interface{} `json:"metadata"`
}

// EncodeCallContext 编码调用上下文（不与 protobuf 的 EncodeRequest 冲突）
type EncodeCallContext struct {
	// 插件上下文
	Context *PluginContext `json:"context"`

	// 待编码的数据
	Data []byte `json:"data"`

	// 额外的元数据
	Metadata map[string]interface{} `json:"metadata"`
}

// NewDecodeCallContext 创建解码调用上下文
func NewDecodeCallContext(pluginName string, isClient bool, payload []byte) *DecodeCallContext {
	return &DecodeCallContext{
		Context:  NewPluginContext(pluginName),
		IsClient: isClient,
		Payload:  payload,
		Metadata: make(map[string]interface{}),
	}
}

// SetMetadata 设置元数据
func (dc *DecodeCallContext) SetMetadata(key string, value interface{}) *DecodeCallContext {
	if dc.Metadata == nil {
		dc.Metadata = make(map[string]interface{})
	}
	dc.Metadata[key] = value
	return dc
}

// SetPluginName 设置插件名称
func (dc *DecodeCallContext) SetPluginName(pluginName string) *DecodeCallContext {
	dc.Context.PluginName = pluginName
	return dc
}

// NewEncodeCallContext 创建编码调用上下文
func NewEncodeCallContext(pluginName string, data []byte) *EncodeCallContext {
	return &EncodeCallContext{
		Context:  NewPluginContext(pluginName),
		Data:     data,
		Metadata: make(map[string]interface{}),
	}
}

// SetMetadata 设置元数据（编码上下文）
func (ec *EncodeCallContext) SetMetadata(key string, value interface{}) *EncodeCallContext {
	if ec.Metadata == nil {
		ec.Metadata = make(map[string]interface{})
	}
	ec.Metadata[key] = value
	return ec
}

// SetPluginName 设置插件名称
func (ec *EncodeCallContext) SetPluginName(pluginName string) *EncodeCallContext {
	ec.Context.PluginName = pluginName
	return ec
}
