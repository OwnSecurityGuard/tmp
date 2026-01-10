package compat

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"proxy-system-backend/internal/types"
)

func TestCompatibilityLayer(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test-compatibility")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 初始化兼容的插件服务
	pluginConfig := types.PluginManagerConfig{
		PluginDir: tempDir,
		MaxSize:   50 * 1024 * 1024, // 50MB
	}
	compatPluginService := NewPluginService(pluginConfig)
	if err := compatPluginService.Initialize(); err != nil {
		t.Fatalf("Failed to initialize plugin service: %v", err)
	}

	// 初始化兼容的WebSocket服务
	wsService := NewWebSocketService()

	// 创建兼容的Shadowsocks服务
	shadowsocksService := NewShadowsocksService(wsService, compatPluginService)

	// 测试配置
	config := types.ServerConfig{
		ListenAddr: "127.0.0.1:0", // 使用随机端口
		Password:   "test-password",
		Method:     "aes-256-gcm",
	}

	// 测试Start方法不应panic
	if err := shadowsocksService.Start(config); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer shadowsocksService.Stop()

	// 验证服务已启动
	time.Sleep(100 * time.Millisecond) // 等待服务启动

	// 测试回调方法不应panic
	shadowsocksService.OnClientConnected()
	shadowsocksService.OnClientDisconnected()

	testData := []byte("test data")
	shadowsocksService.OnPacketReceived("C→S", testData, len(testData), "127.0.0.1:12345", "192.168.1.100:80")
}

func TestFilterRulesAPI(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test-filter-api")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 初始化服务
	pluginConfig := types.PluginManagerConfig{
		PluginDir: tempDir,
		MaxSize:   50 * 1024 * 1024,
	}
	pluginService := NewPluginService(pluginConfig)
	if err := pluginService.Initialize(); err != nil {
		t.Fatalf("Failed to initialize plugin service: %v", err)
	}

	wsService := NewWebSocketService()
	shadowsocksService := NewShadowsocksService(wsService, pluginService)

	config := types.ServerConfig{
		ListenAddr: "127.0.0.1:0",
		Password:   "test-password",
		Method:     "aes-256-gcm",
	}

	if err := shadowsocksService.Start(config); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer shadowsocksService.Stop()

	// 创建HTTP处理器
	// router := handler.SetupRoutes(shadowsocksService, pluginService) // 暂时注释掉，避免循环导入
	router := http.NewServeMux()

	// 测试添加过滤规则
	rule := types.FilterRule{
		ID:          1,
		Name:        "Test Rule",
		Action:      types.FilterActionAllow,
		Direction:   types.FilterDirectionIn,
		SourceIPs:   []types.IPAddress{{IP: "192.168.1.0", Mask: 24}},
		SourcePorts: []types.PortRange{{Min: 1, Max: 60034}},
		DestIPs:     []types.IPAddress{},
		DestPorts:   []types.PortRange{{Min: 8080, Max: 8080}},
		Enabled:     true,
		Priority:    100,
		Description: "测试规则",
	}

	ruleJSON, _ := json.Marshal(rule)
	req, _ := http.NewRequest("POST", "/api/filter/rules", bytes.NewBuffer(ruleJSON))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, status)
	}

	// 测试获取过滤规则
	req, _ = http.NewRequest("GET", "/api/filter/rules", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, status)
	}

	var rules []types.FilterRule
	if err := json.Unmarshal(rr.Body.Bytes(), &rules); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	// 测试更新过滤规则
	rule.Name = "Updated Test Rule"
	ruleJSON, _ = json.Marshal(rule)
	req, _ = http.NewRequest("PUT", "/api/filter/rules/1", bytes.NewBuffer(ruleJSON))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, status)
	}

	// 测试删除过滤规则
	req, _ = http.NewRequest("DELETE", "/api/filter/rules/1", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, status)
	}

	// 验证规则已删除
	req, _ = http.NewRequest("GET", "/api/filter/rules", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, status)
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &rules); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(rules) != 0 {
		t.Fatalf("Expected 0 rules, got %d", len(rules))
	}
}

func TestPluginServiceAPI(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test-plugin-api")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 初始化服务
	pluginConfig := types.PluginManagerConfig{
		PluginDir: tempDir,
		MaxSize:   50 * 1024 * 1024,
	}
	pluginService := NewPluginService(pluginConfig)
	if err := pluginService.Initialize(); err != nil {
		t.Fatalf("Failed to initialize plugin service: %v", err)
	}

	wsService := NewWebSocketService()
	shadowsocksService := NewShadowsocksService(wsService, pluginService)

	config := types.ServerConfig{
		ListenAddr: "127.0.0.1:0",
		Password:   "test-password",
		Method:     "aes-256-gcm",
	}

	if err := shadowsocksService.Start(config); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer shadowsocksService.Stop()

	// 创建HTTP处理器
	// router := handler.SetupRoutes(shadowsocksService, pluginService) // 暂时注释掉，避免循环导入
	router := http.NewServeMux()

	// 创建测试插件文件
	pluginContent := []byte("test plugin content")
	pluginPath := filepath.Join(tempDir, "test-plugin.txt")
	if err := os.WriteFile(pluginPath, pluginContent, 0644); err != nil {
		t.Fatalf("Failed to create test plugin file: %v", err)
	}

	// 测试获取插件列表
	req, _ := http.NewRequest("GET", "/api/plugins", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, status)
	}

	var plugins []types.PluginInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &plugins); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(plugins) != 1 {
		t.Fatalf("Expected 1 plugin, got %d", len(plugins))
	}

	if plugins[0].Name != "test-plugin.txt" {
		t.Fatalf("Expected plugin name 'test-plugin.txt', got %s", plugins[0].Name)
	}
}

func TestDatabasePersistence(t *testing.T) {
	// 创建两个临时目录模拟重启
	tempDir1, err := os.MkdirTemp("", "test-db-persistence-1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir1)

	tempDir2, err := os.MkdirTemp("", "test-db-persistence-2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir2)

	// 第一步：创建服务并添加规则
	pluginConfig1 := types.PluginManagerConfig{
		PluginDir: tempDir1,
		MaxSize:   50 * 1024 * 1024,
	}
	pluginService1 := NewPluginService(pluginConfig1)
	if err := pluginService1.Initialize(); err != nil {
		t.Fatalf("Failed to initialize plugin service: %v", err)
	}

	// 直接访问底层FilterService以添加规则
	// 注意：这只是为了测试目的，实际应用中应通过API或兼容层操作
	// 通过反射或其他方式访问内部服务
	// 这里我们简单地使用数据库直接操作来模拟
	dbPath1 := filepath.Join(tempDir1, "filter.db")
	db1, err := sql.Open("sqlite", dbPath1+"?_journal_mode=WAL")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db1.Close()

	// 创建表
	if _, err := db1.Exec(`
		CREATE TABLE IF NOT EXISTS filter_rules (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			action INTEGER NOT NULL,
			direction INTEGER NOT NULL,
			source_ips TEXT NOT NULL,
			source_ports TEXT NOT NULL,
			dest_ips TEXT NOT NULL,
			dest_ports TEXT NOT NULL,
			enabled INTEGER NOT NULL,
			priority INTEGER NOT NULL,
			description TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 添加测试规则
	rule := types.FilterRule{
		ID:          1,
		Name:        "Test Rule",
		Action:      types.FilterActionAllow,
		Direction:   types.FilterDirectionIn,
		SourceIPs:   []types.IPAddress{{IP: "192.168.1.0", Mask: 24}},
		SourcePorts: []types.PortRange{{Min: 1, Max: 60034}},
		DestIPs:     []types.IPAddress{},
		DestPorts:   []types.PortRange{{Min: 8080, Max: 8080}},
		Enabled:     true,
		Priority:    100,
		Description: "测试规则",
	}

	// 序列化JSON字段
	sourceIPsJSON, _ := json.Marshal(rule.SourceIPs)
	sourcePortsJSON, _ := json.Marshal(rule.SourcePorts)
	destIPsJSON, _ := json.Marshal(rule.DestIPs)
	destPortsJSON, _ := json.Marshal(rule.DestPorts)

	now := time.Now().Format(time.RFC3339)

	// 插入规则
	if _, err := db1.Exec(`
		INSERT INTO filter_rules (
			id, name, action, direction, source_ips, source_ports, 
			dest_ips, dest_ports, enabled, priority, description, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		rule.ID, rule.Name, rule.Action, rule.Direction,
		sourceIPsJSON, sourcePortsJSON, destIPsJSON, destPortsJSON,
		rule.Enabled, rule.Priority, rule.Description, now, now,
	); err != nil {
		t.Fatalf("Failed to insert filter rule: %v", err)
	}

	// 第二步：创建新服务并验证规则是否加载
	pluginConfig2 := types.PluginManagerConfig{
		PluginDir: tempDir2,
		MaxSize:   50 * 1024 * 1024,
	}
	pluginService2 := NewPluginService(pluginConfig2)
	if err := pluginService2.Initialize(); err != nil {
		t.Fatalf("Failed to initialize plugin service: %v", err)
	}

	// 复制数据库文件到新目录
	dbPath2 := filepath.Join(tempDir2, "filter.db")
	dbData, err := os.ReadFile(dbPath1)
	if err != nil {
		t.Fatalf("Failed to read database file: %v", err)
	}
	if err := os.WriteFile(dbPath2, dbData, 0644); err != nil {
		t.Fatalf("Failed to write database file: %v", err)
	}

	// 重新初始化插件服务以加载数据库
	pluginService2 = NewPluginService(pluginConfig2)
	if err := pluginService2.Initialize(); err != nil {
		t.Fatalf("Failed to re-initialize plugin service: %v", err)
	}

	wsService := NewWebSocketService()
	shadowsocksService := NewShadowsocksService(wsService, pluginService2)

	config := types.ServerConfig{
		ListenAddr: "127.0.0.1:0",
		Password:   "test-password",
		Method:     "aes-256-gcm",
	}

	if err := shadowsocksService.Start(config); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer shadowsocksService.Stop()

	// 等待服务加载规则
	time.Sleep(100 * time.Millisecond)

	// 验证规则已加载
	// 这里我们无法直接访问内部PacketFilter，所以通过API验证
	// router := handler.SetupRoutes(shadowsocksService, pluginService2) // 暂时注释掉，避免循环导入
	router := http.NewServeMux()

	req, _ := http.NewRequest("GET", "/api/filter/rules", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, status)
	}

	var rules []types.FilterRule
	if err := json.Unmarshal(rr.Body.Bytes(), &rules); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	if rules[0].ID != 1 {
		t.Fatalf("Expected rule ID 1, got %d", rules[0].ID)
	}

	if rules[0].Name != "Test Rule" {
		t.Fatalf("Expected rule name 'Test Rule', got %s", rules[0].Name)
	}
}
