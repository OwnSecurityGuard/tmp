package plugin

import (
	"context"
	"fmt"
	"time"
)

// PluginInvoker 插件调用器
type PluginInvoker struct {
	manager *Manager
}

// NewPluginInvoker 创建插件调用器
func NewPluginInvoker(manager *Manager) *PluginInvoker {
	return &PluginInvoker{
		manager: manager,
	}
}

// InvokeDecode 调用解码插件
func (inv *PluginInvoker) InvokeDecode(req *DecodeCallContext) (*DecodeResult, error) {
	if inv.manager == nil {
		return nil, fmt.Errorf("plugin manager is nil")
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(req.Context.Context, req.Context.Timeout)
	defer cancel()

	// 将 Go context 注入到请求中
	req.Context.Context = ctx

	// 如果启用了重试
	if req.Context.EnableRetry {
		return inv.invokeWithRetry(req)
	}

	// 单次调用
	return inv.doDecode(req)
}

// InvokeEncode 调用编码插件
func (inv *PluginInvoker) InvokeEncode(req *EncodeCallContext) ([]byte, error) {
	if inv.manager == nil {
		return nil, fmt.Errorf("plugin manager is nil")
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(req.Context.Context, req.Context.Timeout)
	defer cancel()

	// 将 Go context 注入到请求中
	req.Context.Context = ctx

	// 直接调用编码
	return inv.doEncode(req)
}

// doDecode 执行解码
func (inv *PluginInvoker) doDecode(req *DecodeCallContext) (*DecodeResult, error) {
	pluginName := req.Context.PluginName

	// 检查插件是否已加载
	_, ok := inv.manager.Get(pluginName)
	if !ok {
		return nil, fmt.Errorf("plugin '%s' is not loaded", pluginName)
	}

	// 记录调用开始
	if req.Context.Verbose {
		fmt.Printf("[Plugin] Starting decode call: %s\n", req.Context)
	}

	// 执行解码
	startTime := time.Now()
	result, err := inv.manager.Decode(pluginName, req.IsClient, req.Payload)
	duration := time.Since(startTime)

	if err != nil {
		if req.Context.Verbose {
			fmt.Printf("[Plugin] Decode call failed after %v: %v\n", duration, err)
		}
		return nil, err
	}

	if req.Context.Verbose {
		fmt.Printf("[Plugin] Decode call completed in %v\n", duration)
	}

	return result, nil
}

// doEncode 执行编码
func (inv *PluginInvoker) doEncode(req *EncodeCallContext) ([]byte, error) {
	pluginName := req.Context.PluginName

	// 检查插件是否已加载
	_, ok := inv.manager.Get(pluginName)
	if !ok {
		return nil, fmt.Errorf("plugin '%s' is not loaded", pluginName)
	}

	// 记录调用开始
	if req.Context.Verbose {
		fmt.Printf("[Plugin] Starting encode call: %s\n", req.Context)
	}

	// 执行编码
	startTime := time.Now()
	// 注意：Manager 目前没有直接暴露 Encode 方法，这里需要扩展
	// 暂时返回错误
	duration := time.Since(startTime)

	if req.Context.Verbose {
		fmt.Printf("[Plugin] Encode call completed in %v\n", duration)
	}

	return nil, fmt.Errorf("encode method not implemented in manager")
}

// invokeWithRetry 带重试的解码调用
func (inv *PluginInvoker) invokeWithRetry(req *DecodeCallContext) (*DecodeResult, error) {
	maxRetries := req.Context.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if req.Context.Verbose {
				fmt.Printf("[Plugin] Retry attempt %d/%d for plugin '%s'\n",
					attempt, maxRetries, req.Context.PluginName)
			}

			// 等待重试延迟
			time.Sleep(req.Context.RetryDelay)
		}

		result, err := inv.doDecode(req)
		if err == nil {
			if attempt > 0 && req.Context.Verbose {
				fmt.Printf("[Plugin] Retry succeeded on attempt %d\n", attempt)
			}
			return result, nil
		}

		lastErr = err

		// 检查是否应该重试
		if !shouldRetry(err) {
			break
		}
	}

	return nil, fmt.Errorf("decode failed after %d attempts: %w", maxRetries+1, lastErr)
}

// shouldRetry 判断错误是否可以重试
func shouldRetry(err error) bool {
	// 可以根据错误类型判断是否应该重试
	// 例如：超时、临时性错误等
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// 超时错误可以重试
	if contains(errMsg, "timeout") || contains(errMsg, "deadline exceeded") {
		return true
	}

	// 连接错误可以重试
	if contains(errMsg, "connection") || contains(errMsg, "network") {
		return true
	}

	return false
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// InvokeWithDefaultBehavior 使用默认行为调用解码
// 这是一个简化的接口，向后兼容
func InvokeWithDefaultBehavior(manager *Manager, pluginName string, isClient bool, payload []byte) (*DecodeResult, error) {
	invoker := NewPluginInvoker(manager)
	req := NewDecodeCallContext(pluginName, isClient, payload)
	return invoker.InvokeDecode(req)
}
