# 插件系统配置说明

## 概述

本系统已将插件系统从硬编码参数改为动态配置方式，支持通过配置文件和环境变量灵活管理插件。同时提供了完整的流量处理插件调用机制，支持参数化调用和运行时配置。

## 配置文件

### 主配置文件：`plugin_config.json`

默认配置文件，用于生产环境。

```json
{
  "handshake_config": {
    "protocol_version": 1,
    "magic_cookie_key": "GAME_PROTOCOL_PLUGIN",
    "magic_cookie_value": "hello"
  },
  "plugin_dir": "./data/plugins",
  "manager": {
    "grpc_dial_options": [],
    "allowed_protocols": ["grpc"],
    "plugin_mapping": {"protocol": "protocol"},
    "max_concurrent_plugins": 10,
    "load_timeout": 30,
    "auto_load_plugins": []
  },
  "proxy": {
    "traffic_hook": {
      "decoder_plugin": "",
      "enabled": false,
      "fallback_behavior": "pass",
      "log_decode_errors": true,
      "timeout_ms": 5000
    },
    "auto_load_plugins": []
  },
  "debug": {
    "enabled": false,
    "default_plugin_name": "",
    "test_data": "",
    "verbose_logging": false
  }
}
```

### 调试配置文件：`plugin_config_debug.json`

调试模式配置文件，包含详细的日志和测试数据。

```json
{
  "debug": {
    "enabled": true,
    "default_plugin_name": "test",
    "test_data": "hhhhhhhhhhhhhhhhhhh",
    "verbose_logging": true
  },
  "proxy": {
    "traffic_hook": {
      "decoder_plugin": "test",
      "enabled": true,
      "fallback_behavior": "fallback",
      "log_decode_errors": true,
      "timeout_ms": 5000
    }
  }
}
```

## 配置项说明

### 1. handshake_config（握手配置）

| 字段 | 类型 | 说明 |
|------|------|------|
| protocol_version | uint | 协议版本号 |
| magic_cookie_key | string | 魔术密钥键名 |
| magic_cookie_value | string | 魔术密钥值 |

### 2. manager（插件管理器配置）

| 字段 | 类型 | 说明 |
|------|------|------|
| grpc_dial_options | []string | GRPC拨号选项列表 |
| allowed_protocols | []string | 允许的协议类型（grpc/netrpc） |
| plugin_mapping | map[string]string | 插件名称映射 |
| max_concurrent_plugins | int | 最大并发插件数量 |
| load_timeout | int | 插件加载超时时间（秒） |
| auto_load_plugins | []string | 启动时自动加载的插件列表 |

### 3. proxy（代理配置）

#### traffic_hook（流量钩子配置）

| 字段 | 类型 | 说明 |
|------|------|------|
| decoder_plugin | string | 流量解码插件名称 |
| enabled | bool | 是否启用流量钩子 |
| fallback_behavior | string | 解码失败时的回退行为（pass/drop/fallback） |
| log_decode_errors | bool | 是否记录解码错误 |
| timeout_ms | int | 插件调用超时时间（毫秒） |

#### auto_load_plugins（代理自动加载插件）

插件列表，代理启动时自动加载。

### 4. debug（调试配置）

| 字段 | 类型 | 说明 |
|------|------|------|
| enabled | bool | 是否启用调试模式 |
| default_plugin_name | string | 调试模式下的默认插件名 |
| test_data | string | 调试模式下的测试数据 |
| verbose_logging | bool | 是否打印详细日志 |

## 使用方法

### 1. 基础使用

在启动服务时，插件系统会自动加载默认配置文件：

```go
import "proxy-system-backend/internal/modules/plugin"

// 使用默认配置初始化
manager, err := plugin.InitializePluginSystem("")
if err != nil {
    log.Fatal(err)
}
```

### 2. 指定配置文件

```go
manager, err := plugin.InitializePluginSystem("/path/to/config.json")
if err != nil {
    log.Fatal(err)
}
```

### 3. 环境变量切换调试模式

设置环境变量 `PLUGIN_DEBUG_MODE` 为 `true` 或 `1` 可启用调试模式：

```bash
# Linux/Mac
export PLUGIN_DEBUG_MODE=true
./server

# Windows
set PLUGIN_DEBUG_MODE=true
server.exe
```

### 4. 运行时加载配置

```go
// 加载配置
config, err := plugin.LoadConfig("./plugin_config.json")
if err != nil {
    log.Fatal(err)
}

// 使用配置创建管理器
manager := plugin.NewManager(config)

// 或更新现有管理器的配置
err = plugin.InitializePluginSystemWithManager(manager, "./plugin_config.json")
```

## 流量钩子插件调用

### 基本使用

流量钩子插件用于在数据包处理时自动调用配置的解码插件。配置后，系统会自动在 `OnPacket` 中调用插件。

```json
{
  "proxy": {
    "traffic_hook": {
      "decoder_plugin": "my_decoder",
      "enabled": true
    }
  }
}
```

### 回退行为

当插件解码失败时，系统会根据配置的回退行为处理：

1. **pass** - 传递原始数据包，继续处理
2. **drop** - 丢弃数据包
3. **fallback** - 尝试使用备用插件，最终回退到传递原始数据

### 运行时参数传递

使用 `PluginInvoker` 进行带上下文的插件调用：

```go
import (
    "proxy-system-backend/internal/modules/plugin"
    "time"
)

// 创建调用器
invoker := plugin.NewPluginInvoker(manager)

// 创建解码请求
req := plugin.NewDecodeRequest("plugin_name", true, payload)

// 设置自定义参数
req.SetParam("proxy_id", "proxy-1")
req.SetParam("custom_key", "custom_value")

// 设置超时和重试
req.Context.
    SetTimeout(5 * time.Second).
    EnableRetryWithConfig(3, 100*time.Millisecond).
    SetVerbose(true)

// 调用插件
result, err := invoker.InvokeDecode(req)
```

### 插件上下文

`PluginContext` 提供了丰富的参数传递机制：

```go
ctx := plugin.NewPluginContext("plugin_name")

// 设置参数
ctx.SetParam("key", "value")

// 获取参数
if val, ok := ctx.GetParam("key"); ok {
    fmt.Println(val)
}

// 配置调用选项
ctx.SetTimeout(3 * time.Second)
ctx.SetVerbose(true)
ctx.EnableRetryWithConfig(5, 200*time.Millisecond)
```

## 配置验证

### 检查流量钩子状态

```go
if plugin.IsTrafficHookEnabled() {
    fmt.Println("Traffic hook is enabled")
    decoderPlugin := plugin.GetTrafficHookDecoderPlugin()
    fmt.Printf("Using decoder: %s\n", decoderPlugin)
}
```

### 检查是否启用调试模式

```go
if plugin.IsEnabled() {
    fmt.Println("Debug mode is enabled")
    defaultPlugin := plugin.GetDefaultPluginName()
    testData := plugin.GetTestData()
}
```

### 获取当前配置

```go
config := plugin.GetConfig()
fmt.Printf("Plugin dir: %s\n", config.PluginDir)
fmt.Printf("Debug mode: %v\n", config.Debug.Enabled)
fmt.Printf("Traffic hook enabled: %v\n", config.Proxy.TrafficHook.Enabled)
```

## 兼容性

系统保留了向后兼容性：

1. **Handshake 配置**：原有的 `Handshake` 变量仍可使用，但建议使用 `GetHandshakeConfig()` 方法
2. **插件加载**：原有的加载方式仍然支持，但推荐使用配置文件管理
3. **向后兼容方法**：`InvokeWithDefaultBehavior` 提供了简化的调用接口

## 迁移指南

### 从硬编码迁移到配置文件

**之前的代码：**
```go
// proxy_traffic_hook.go - 硬编码插件名称
data, _ := p.Decode("test", true, ctx.Payload)
```

**迁移后的代码：**
```go
// 自动使用配置的插件
decoderPlugin := h.getDecoderPlugin()
if decoderPlugin != "" {
    data, err := h.decodeWithPlugin(decoderPlugin, ctx)
    // ...
}
```

### 使用新的调用器

**之前的代码：**
```go
data, err := manager.Decode("plugin_name", true, payload)
```

**迁移后的代码：**
```go
invoker := plugin.NewPluginInvoker(manager)
req := plugin.NewDecodeRequest("plugin_name", true, payload)
result, err := invoker.InvokeDecode(req)
```

## 调试模式

启用调试模式后，系统会：

1. 自动加载配置的默认插件
2. 执行测试数据调用
3. 打印详细日志
4. 在插件加载失败时显示详细信息

切换调试模式：

1. 修改 `plugin_config.json` 中的 `debug.enabled` 为 `true`
2. 或使用 `plugin_config_debug.json` 配置文件
3. 或设置环境变量 `PLUGIN_DEBUG_MODE=true`

## 故障排查

### 插件加载失败

1. 检查配置文件路径是否正确
2. 检查插件文件是否存在且有执行权限
3. 启用详细日志查看详细错误信息：
   ```json
   "debug": {
     "verbose_logging": true
   }
   ```

### 流量解码失败

1. 检查 `traffic_hook.enabled` 是否为 `true`
2. 检查 `decoder_plugin` 是否正确
3. 检查插件是否已加载
4. 查看回退行为配置
5. 启用 `log_decode_errors` 查看详细错误

### 握手失败

检查握手配置是否与插件端一致：
- `magic_cookie_key`
- `magic_cookie_value`
- `protocol_version`

### 配置文件不生效

1. 确认配置文件格式正确（JSON）
2. 检查配置文件路径
3. 使用 `LoadConfig()` 直接加载测试

## 最佳实践

1. **生产环境**：使用 `plugin_config.json`，禁用调试模式
2. **开发环境**：使用 `plugin_config_debug.json` 或环境变量
3. **敏感配置**：不要将真实密钥提交到版本控制
4. **备份配置**：保留配置文件的备份
5. **版本管理**：不同环境使用不同的配置文件
6. **流量钩子**：在生产环境中设置合理的超时时间和回退行为
7. **参数传递**：使用 `PluginContext` 传递自定义参数，便于追踪和调试

## 高级功能

### 重试机制

```go
req := plugin.NewDecodeRequest("plugin_name", true, payload)
req.Context.EnableRetryWithConfig(
    3,              // 最大重试次数
    100*time.Millisecond, // 重试延迟
)
```

### 自定义超时

```go
req.Context.SetTimeout(10 * time.Second)
```

### 元数据传递

```go
req.SetMetadata("proxy_id", "proxy-1")
req.SetMetadata("timestamp", time.Now().Unix())
```

### 详细日志

```go
req.Context.SetVerbose(true)
```

## API 参考

### 配置相关函数

- `LoadConfig(path string) (*Config, error)` - 加载配置文件
- `SaveConfig(path string, config *Config) error` - 保存配置文件
- `GetConfig() *Config` - 获取当前配置
- `UpdateConfig(config *Config)` - 更新配置

### 流量钩子相关函数

- `IsTrafficHookEnabled() bool` - 检查流量钩子是否启用
- `GetTrafficHookDecoderPlugin() string` - 获取解码插件名称
- `GetTrafficHookTimeout() int` - 获取超时时间
- `ShouldLogDecodeErrors() bool` - 检查是否记录错误
- `GetFallbackBehavior() string` - 获取回退行为

### 调试相关函数

- `IsEnabled() bool` - 检查调试模式是否启用
- `IsDebugEnabled() bool` - 检查调试模式是否启用（别名）
- `GetDefaultPluginName() string` - 获取默认插件名
- `GetTestData() string` - 获取测试数据

### 插件调用相关类型

- `PluginContext` - 插件调用上下文
- `DecodeRequest` - 解码请求
- `EncodeRequest` - 编码请求
- `PluginInvoker` - 插件调用器
