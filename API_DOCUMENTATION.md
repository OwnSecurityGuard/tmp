# Shadowsocks 代理管理系统 API 文档

> 版本：v1.0.0  
> 最后更新：2025-01-17  
> 基础URL：`http://localhost:8081/api`

## 目录

1. [概述](#概述)
2. [认证说明](#认证说明)
3. [接口清单](#接口清单)
4. [代理服务接口](#代理服务接口)
5. [插件管理接口](#插件管理接口)
6. [WebSocket实时通信](#websocket实时通信)
7. [错误码说明](#错误码说明)
8. [调用示例](#调用示例)
9. [测试环境](#测试环境)
10. [版本控制](#版本控制)

## 概述

Shadowsocks代理管理系统提供了一套完整的RESTful API接口，用于管理代理服务、插件和实时监控流量数据。系统基于Gin框架构建，使用WebSocket进行实时数据推送。

### 技术特性

- **协议**：HTTP/1.1 + WebSocket
- **数据格式**：JSON
- **跨域支持**：CORS已启用
- **心跳检测**：30秒WebSocket心跳
- **错误处理**：统一JSON错误响应

## 认证说明

> **⚠️ 重要提示**：当前系统**未实现认证机制**，所有API端点公开可访问。建议在生产环境部署时添加JWT或API Key认证。

### 建议的认证方案（待实现）

```
// 请求头格式
Authorization: Bearer <JWT_TOKEN>
// 或
X-API-Key: <API_KEY>
```

## 接口清单

### HTTP REST API

| 模块 | 方法 | 路径 | 功能描述 |
|------|------|------|----------|
| **系统** | GET | `/health` | 健康检查 |
| **代理** | POST | `/api/proxy/start` | 启动代理服务 |
| **插件** | GET | `/api/plugins` | 获取插件列表 |
| **插件** | POST | `/api/plugins` | 注册插件 |
| **插件** | GET | `/api/plugins/:name` | 获取插件详情 |
| **插件** | POST | `/api/plugins/:name/load` | 加载插件 |
| **插件** | POST | `/api/plugins/:name/unload` | 卸载插件 |
| **插件** | POST | `/api/plugins/upload` | 上传插件 |
| **实时** | GET | `/api/ws` | WebSocket连接 |

### WebSocket事件类型

| 事件类型 | 描述 | 数据格式 |
|----------|------|----------|
| `welcome` | 连接成功欢迎消息 | `{client_id, token}` |
| `proxy_started` | 代理启动 | `{proxy_id, listen_addr}` |
| `proxy_stopped` | 代理停止 | `{proxy_id}` |
| `plugin_loaded` | 插件加载 | `{plugin_name}` |
| `plugin_unloaded` | 插件卸载 | `{plugin_name}` |
| `traffic` | 流量数据 | `{proxy_id, conn_id, payload}` |
| `parsed` | 解析数据 | 插件解析后的数据 |

## 代理服务接口

### 启动代理服务

**接口信息**

- **URL**：`/api/proxy/start`
- **方法**：`POST`
- **Content-Type**：`application/json`

**请求参数**

```json
{
  "block_ips": ["192.168.1.100", "10.0.0.0/24"],  // 可选，阻止的IP地址列表
  "block_ports": ["8080", "9000-9100"],         // 可选，阻止的端口或端口范围
  "plugin_name": "custom_decoder"               // 可选，使用的插件名称
}
```

**参数说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| block_ips | array[string] | 否 | 阻止的IP地址列表，支持单个IP或CIDR格式 |
| block_ports | array[string] | 否 | 阻止的端口列表，支持单个端口或范围（如"9000-9100"） |
| plugin_name | string | 否 | 流量解码插件名称，需要在插件管理中预先注册 |

**成功响应**

```json
{
  "success": true,
  "message": "proxy started successfully",
  "data": {
    "id": "proxy-xxx",
    "listen_addr": ":8388",
    "status": "running",
    "start_time": 1234567890
  }
}
```

**错误响应**

```json
{
  "error": "port already in use"
}

{
  "error": "invalid cipher method"
}

{
  "error": "plugin not found: custom_decoder"
}
```

**调用示例**

```javascript
// 基础调用（无过滤）
fetch('http://localhost:8081/api/proxy/start', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({})
});

// 带IP和端口过滤的调用
fetch('http://localhost:8081/api/proxy/start', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    block_ips: ['192.168.1.100', '10.0.0.0/24'],
    block_ports: ['8080', '9000-9100'],
    plugin_name: 'custom_decoder'
  })
});
```

## 插件管理接口

### 获取插件列表

**接口信息**

- **URL**：`/api/plugins`
- **方法**：`GET`

**请求参数**

无

**成功响应**

```json
{
  "plugins": [
    {
      "name": "test-plugin",
      "path": "E:\\reply\\backend\\data\\plugins\\test-plugin\\plugin.exe",
      "size": 1024000,
      "created_at": 1234567890,
      "modified_at": 1234567890,
      "status": "active",
      "description": "测试插件"
    },
    {
      "name": "game-protocol",
      "path": "E:\\reply\\backend\\data\\plugins\\game-protocol\\plugin.exe",
      "size": 2048000,
      "created_at": 1234567890,
      "modified_at": 1234567890,
      "status": "inactive",
      "description": "游戏协议解析插件"
    }
  ]
}
```

**数据说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 插件名称（唯一标识） |
| path | string | 插件文件完整路径 |
| size | int64 | 插件文件大小（字节） |
| created_at | int | 创建时间（Unix时间戳） |
| modified_at | int | 最后修改时间（Unix时间戳） |
| status | string | 插件状态（active/inactive/error/running） |
| description | string | 插件描述信息 |

### 注册插件

**接口信息**

- **URL**：`/api/plugins`
- **方法**：`POST`
- **Content-Type**：`application/json`

**请求参数**

```json
{
  "name": "game-protocol",
  "path": "./plugins/game-protocol.exe",
  "description": "游戏协议解析插件"
}
```

**参数说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 插件名称（唯一） |
| path | string | 是 | 插件文件路径 |
| description | string | 否 | 插件描述 |

**成功响应**

```json
{
  "success": true,
  "message": "plugin registered successfully"
}
```

**错误响应**

```json
{
  "error": "plugin 'game-protocol' already exists"
}

{
  "error": "plugin file not found: ./plugins/game-protocol.exe"
}
```

### 获取插件详情

**接口信息**

- **URL**：`/api/plugins/:name`
- **方法**：`GET`

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| name | string | 插件名称 |

**成功响应**

```json
{
  "name": "game-protocol",
  "path": "E:\\reply\\backend\\data\\plugins\\game-protocol\\plugin.exe",
  "size": 2048000,
  "created_at": 1234567890,
  "modified_at": 1234567890,
  "status": "running",
  "description": "游戏协议解析插件"
}
```

**错误响应**

```json
{
  "error": "plugin 'game-protocol' not found"
}
```

### 加载插件

**接口信息**

- **URL**：`/api/plugins/:name/load`
- **方法**：`POST`

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| name | string | 插件名称 |

**请求体**

空或可选配置（根据插件需求）

**成功响应**

```json
{
  "success": true,
  "message": "plugin loaded successfully",
  "data": {
    "name": "game-protocol",
    "status": "running",
    "loaded_at": 1234567890
  }
}
```

**错误响应**

```json
{
  "error": "plugin 'game-protocol' not found"
}

{
  "error": "failed to load plugin: handshake failed"
}
```

### 卸载插件

**接口信息**

- **URL**：`/api/plugins/:name/unload`
- **方法**：`POST`

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| name | string | 插件名称 |

**成功响应**

```json
{
  "success": true,
  "message": "plugin unloaded successfully"
}
```

**错误响应**

```json
{
  "error": "plugin 'game-protocol' not loaded"
}
```

### 上传插件

**接口信息**

- **URL**：`/api/plugins/upload`
- **方法**：`POST`
- **Content-Type**：`multipart/form-data`

**请求参数**

```javascript
const formData = new FormData();
formData.append('name', 'new-plugin');           // 插件名称
formData.append('description', '描述信息');       // 插件描述
formData.append('file', fileInput.files[0]);     // 插件文件（.exe/.so/.bin）
```

**参数说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 插件名称，如果不提供则使用文件名 |
| description | string | 否 | 插件描述 |
| file | file | 是 | 插件可执行文件 |

**成功响应**

```json
{
  "success": true,
  "path": "E:\\reply\\backend\\data\\plugins\\new-plugin\\plugin.exe"
}
```

**错误响应**

```json
{
  "error": "file is required"
}

{
  "error": "mkdir failed: permission denied"
}

{
  "error": "save file failed: disk full"
}
```

**注意事项**

- 插件文件会保存在 `./data/plugins/{name}/{filename}`
- Linux/macOS 系统会自动设置可执行权限
- Windows 系统需要确保文件本身是可执行格式（.exe）

## WebSocket实时通信

### 建立WebSocket连接

**接口信息**

- **URL**：`ws://localhost:8081/api/ws`
- **协议**：WebSocket
- **心跳间隔**：30秒

**连接示例**

```javascript
const ws = new WebSocket('ws://localhost:8081/api/ws');

// 连接成功
ws.onopen = function(event) {
    console.log('WebSocket connected');
};

// 接收消息
ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
    handleEvent(data);
};

// 连接关闭
ws.onclose = function(event) {
    console.log('WebSocket closed');
    // 实现重连逻辑
};

// 错误处理
ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};
```

### WebSocket消息格式

#### 欢迎消息（连接成功后立即发送）

```json
{
  "type": "welcome",
  "client_id": "client-123456",
  "token": "ws-token-xxx",
  "message": "连接成功",
  "timestamp": 1234567890
}
```

#### 代理启动事件

```json
{
  "type": "proxy_started",
  "data": {
    "id": "proxy-xxx",
    "listen_addr": ":8388",
    "status": "running",
    "start_time": 1234567890
  },
  "timestamp": 1234567890
}
```

#### 代理停止事件

```json
{
  "type": "proxy_stopped",
  "data": {
    "id": "proxy-xxx"
  },
  "timestamp": 1234567890
}
```

#### 插件加载事件

```json
{
  "type": "plugin_loaded",
  "data": {
    "name": "game-protocol",
    "status": "running",
    "loaded_at": 1234567890
  },
  "timestamp": 1234567890
}
```

#### 插件卸载事件

```json
{
  "type": "plugin_unloaded",
  "data": {
    "name": "game-protocol"
  },
  "timestamp": 1234567890
}
```

#### 流量数据事件

```json
{
  "type": "traffic",
  "data": {
    "proxy_id": "proxy-xxx",
    "conn_id": "conn-123",
    "payload": {
      "conn_id": "conn-123",
      "direction": "out",
      "src_addr": "192.168.1.100:54321",
      "dst_addr": "8.8.8.8:80",
      "src_ip": "192.168.1.100",
      "src_port": 54321,
      "dst_ip": "8.8.8.8",
      "dst_port": 80,
      "payload": "base64_encoded_data",
      "start_at": "2025-01-17T10:30:00Z"
    }
  },
  "timestamp": 1234567890
}
```

#### 解析数据事件（插件处理后）

```json
{
  "type": "parsed",
  "data": {
    "is_client": true,
    "time": 1234567890123,
    "data": {
      "protocol": "http",
      "method": "GET",
      "path": "/index.html",
      "headers": {...}
    }
  },
  "timestamp": 1234567890
}
```

### 心跳机制

服务器每30秒发送一次ping消息，客户端应回复pong以保持连接。

```javascript
// 监听ping并回复pong
ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    if (data.type === 'ping') {
        ws.send(JSON.stringify({type: 'pong'}));
    }
};
```

## 错误码说明

### HTTP状态码

| 状态码 | 含义 | 说明 |
|--------|------|------|
| 200 | OK | 请求成功 |
| 400 | Bad Request | 请求参数错误或业务逻辑错误 |
| 404 | Not Found | 资源不存在 |
| 500 | Internal Server Error | 服务器内部错误 |
| 204 | No Content | 请求成功但无返回内容（CORS预检） |

### 错误响应格式

```json
{
  "error": "错误描述信息"
}
```

### 常见错误码

| 错误场景 | 错误消息 |
|----------|----------|
| 参数验证失败 | `"invalid request body"` 或具体字段错误 |
| 插件已存在 | `"plugin 'xxx' already exists"` |
| 插件未找到 | `"plugin 'xxx' not found"` |
| 插件未加载 | `"plugin 'xxx' not loaded"` |
| 端口被占用 | `"port already in use"` |
| 文件过大 | `"plugin file too large: 52428800 bytes (max: 52428800)"` |
| 数据库错误 | `"failed to query plugins: database error"` |
| 插件加载失败 | `"failed to load plugin: handshake failed"` |
| 无效加密方法 | `"invalid cipher method"` |

## 调用示例

### 完整示例：启动带过滤和插件的代理

```javascript
// 步骤1: 上传插件
async function uploadPlugin() {
    const formData = new FormData();
    formData.append('name', 'game-decoder');
    formData.append('description', '游戏协议解码插件');
    formData.append('file', fileInput.files[0]);

    const response = await fetch('http://localhost:8081/api/plugins/upload', {
        method: 'POST',
        body: formData
    });
    
    return await response.json();
}

// 步骤2: 加载插件
async function loadPlugin(pluginName) {
    const response = await fetch(`http://localhost:8081/api/plugins/${pluginName}/load`, {
        method: 'POST'
    });
    
    return await response.json();
}

// 步骤3: 启动代理
async function startProxy() {
    const response = await fetch('http://localhost:8081/api/proxy/start', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({
            block_ips: ['192.168.1.100', '10.0.0.0/24'],
            block_ports: ['8080', '9000-9100'],
            plugin_name: 'game-decoder'
        })
    });
    
    return await response.json();
}

// 步骤4: 建立WebSocket连接监听事件
function connectWebSocket() {
    const ws = new WebSocket('ws://localhost:8081/api/ws');
    
    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        
        switch(data.type) {
            case 'proxy_started':
                console.log('代理已启动:', data.data.listen_addr);
                break;
            case 'plugin_loaded':
                console.log('插件已加载:', data.data.name);
                break;
            case 'traffic':
                console.log('收到流量数据:', data.data.payload);
                break;
            case 'parsed':
                console.log('解析后的数据:', data.data);
                break;
        }
    };
    
    return ws;
}

// 完整流程
async function setupProxy() {
    try {
        // 上传并加载插件
        await uploadPlugin();
        await loadPlugin('game-decoder');
        
        // 启动代理
        const proxyResult = await startProxy();
        console.log('代理启动成功:', proxyResult.data);
        
        // 连接WebSocket
        const ws = connectWebSocket();
        
        return { proxy: proxyResult.data, ws };
    } catch (error) {
        console.error('设置失败:', error);
    }
}
```

### React Hooks 示例

```typescript
import { useEffect, useState } from 'react';

interface PluginInfo {
    name: string;
    status: string;
    description: string;
}

function usePlugins() {
    const [plugins, setPlugins] = useState<PluginInfo[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetchPlugins();
    }, []);

    const fetchPlugins = async () => {
        try {
            const response = await fetch('http://localhost:8081/api/plugins');
            const result = await response.json();
            setPlugins(result.plugins || []);  // 注意：实际返回是 {plugins: [...]}
        } catch (error) {
            console.error('Failed to fetch plugins:', error);
        } finally {
            setLoading(false);
        }
    };

    const uploadPlugin = async (file: File, name: string, description: string) => {
        const formData = new FormData();
        formData.append('name', name);
        formData.append('description', description);
        formData.append('file', file);

        const response = await fetch('http://localhost:8081/api/plugins/upload', {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            await fetchPlugins(); // 刷新列表
        }

        return response.ok;
    };

    const loadPlugin = async (name: string) => {
        const response = await fetch(`http://localhost:8081/api/plugins/${name}/load`, {
            method: 'POST'
        });
        return response.ok;
    };

    return { plugins, loading, uploadPlugin, loadPlugin };
}

function useWebSocket(onMessage: (data: any) => void) {
    const [connected, setConnected] = useState(false);

    useEffect(() => {
        const ws = new WebSocket('ws://localhost:8081/api/ws');

        ws.onopen = () => {
            setConnected(true);
        };

        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            onMessage(data);
        };

        ws.onclose = () => {
            setConnected(false);
        };

        return () => {
            ws.close();
        };
    }, [onMessage]);

    return { connected };
}
```

### 完整的curl测试示例

#### 1. 健康检查
```bash
curl -X GET http://localhost:8081/health
```

**响应**：
```json
{"status": "ok"}
```

#### 2. 获取插件列表
```bash
curl -X GET http://localhost:8081/api/plugins
```

**响应示例**：
```json
{
  "plugins": [
    {
      "name": "test-plugin",
      "path": "E:\\reply\\backend\\data\\plugins\\test-plugin\\plugin.exe",
      "size": 1024000,
      "created_at": 1737105600,
      "modified_at": 1737105600,
      "status": "active",
      "description": "测试插件"
    }
  ]
}
```

#### 3. 注册插件（已存在文件）
```bash
curl -X POST http://localhost:8081/api/plugins \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-plugin",
    "path": "./data/plugins/my-plugin/plugin.exe"
  }'
```

**响应**：
```json
{"success": true}
```

#### 4. 加载插件
```bash
curl -X POST http://localhost:8081/api/plugins/my-plugin/load
```

**响应**：
```json
{"success": true}
```

#### 5. 启动代理
```bash
curl -X POST http://localhost:8081/api/proxy/start \
  -H "Content-Type: application/json" \
  -d '{}'
```

**响应示例**：
```json
{
  "success": true,
  "message": "proxy started successfully",
  "data": {
    "id": "proxy-abc123",
    "listen_addr": ":8388",
    "status": "running",
    "start_time": 1737105600
  }
}
```

#### 6. 测试WebSocket连接
```bash
# 使用websocat工具
websocat ws://localhost:8081/api/ws

# 或使用浏览器控制台
const ws = new WebSocket('ws://localhost:8081/api/ws');
ws.onmessage = (e) => console.log(JSON.parse(e.data));
```

**连接后收到的第一条消息**：
```json
{
  "type": "welcome",
  "client_id": "client-7h3k9f2n",
  "token": "ws-token-xxx",
  "message": "连接成功",
  "timestamp": 1737105600
}
```

**代理启动事件**：
```json
{
  "type": "proxy_started",
  "data": {
    "id": "proxy-abc123",
    "listen_addr": ":8388",
    "status": "running",
    "start_time": 1737105600
  },
  "timestamp": 1737105601
}
```

## 测试环境

### 环境信息

- **基础URL**：`http://localhost:8081`
- **WebSocket URL**：`ws://localhost:8081/api/ws`
- **CORS策略**：允许 `http://localhost:3000`（开发环境）

### 测试插件文件

可以使用以下路径测试插件功能：
- `./plugins/test-plugin.exe` - 测试插件
- `./plugins/game-protocol.exe` - 游戏协议插件

### 测试配置示例

```json
{
  "block_ips": ["192.168.1.100"],
  "block_ports": ["8080"],
  "plugin_name": "test"
}
```

### 启动测试服务器

```bash
# 从项目根目录
cd cmd/refactor
go run main.go

# 或使用编译后的版本
./server
```

服务器将在 `http://localhost:8081` 启动。

### 健康检查

```bash
curl http://localhost:8081/health
```

响应：
```json
{"status": "ok"}
```

## 版本控制

### 当前版本

- **API版本**：v1.0.0
- **基础路径**：`/api`

### 版本升级策略

当API有重大变更时，将通过以下方式之一实现版本控制：

1. **URL路径版本**：`/api/v2/...`
2. **请求头版本**：`Accept: application/vnd.proxy.v2+json`
3. **查询参数版本**：`?version=2`

### 变更记录

#### v1.0.0 (2025-01-17)
- 初始版本发布
- 支持代理启动/停止
- 支持插件管理（CRUD）
- 支持WebSocket实时通信
- 支持流量过滤（IP/端口）

## 性能与限制

### 插件限制

| 项目 | 限制值 | 说明 |
|------|--------|------|
| 插件文件大小 | 50MB | 单个插件文件最大限制 |
| 最大并发插件 | 10个 | 同时运行的插件数量 |
| 插件加载超时 | 30秒 | 插件加载超时时间 |

### 代理限制

| 项目 | 限制值 | 说明 |
|------|--------|------|
| 监听地址 | 任意端口 | 支持自定义监听地址 |
| 加密方法 | aes-256-gcm等 | 支持标准Shadowsocks加密方法 |
| 并发连接 | 无限制 | 依赖系统资源 |

### API调用限制

> **当前未实现速率限制**，建议后续添加以下策略：

- 插件上传：每分钟最多5次
- 代理启停：每分钟最多10次
- WebSocket连接：每IP最多5个并发连接

## 安全建议

### 生产环境部署

1. **添加认证机制**：实现JWT或API Key认证
2. **HTTPS支持**：配置TLS证书
3. **速率限制**：实现API调用频率限制
4. **输入验证**：加强请求参数验证
5. **日志审计**：记录重要操作日志

### 建议的安全头

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

## 技术支持

### 常见问题

**Q: WebSocket连接断开怎么办？**
A: 实现自动重连机制，监听onclose事件并重新连接。

**Q: 插件加载失败如何处理？**
A: 检查插件文件是否存在，查看服务器日志，确保握手配置匹配。

**Q: 如何调试API调用？**
A: 使用WebSocket连接监听事件，查看服务器控制台日志，检查返回的错误信息。

### 调试技巧

1. **查看服务器日志**：启动服务器时会显示详细日志
2. **WebSocket调试**：使用浏览器开发者工具查看WebSocket消息
3. **cURL测试**：
   ```bash
   # 测试插件列表
   curl http://localhost:8081/api/plugins

   # 测试启动代理
   curl -X POST http://localhost:8081/api/proxy/start \
     -H "Content-Type: application/json" \
     -d '{}'
   ```

## 附录

### Shadowsocks加密方法列表

支持的加密方法：
- aes-256-gcm
- aes-128-gcm
- chacha20-poly1305
- ...（其他标准方法）

### 插件开发指南

插件需要实现以下接口：

```go
type ProtocolPlugin interface {
    Decode(payload []byte, isClient bool) (*DecodeResult, error)
    Encode(data []byte) ([]byte, error)
}
```

详见：[插件开发文档](./PLUGIN_DEVELOPMENT.md)

### 相关文档

- [插件配置文档](./PLUGIN_CONFIG.md) - 插件系统配置说明
- [插件开发指南](./PLUGIN_DEVELOPMENT.md) - 如何开发自定义插件
- [部署文档](./DEPLOYMENT.md) - 生产环境部署指南

---

**文档版本**：v1.0.0  
**最后更新**：2025-01-17  
**维护者**：Shadowsocks代理管理系统开发团队
