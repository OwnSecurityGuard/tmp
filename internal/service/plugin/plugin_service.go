package plugin

import (
	"database/sql"

	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// PluginService 插件管理服务
type PluginService struct {
	config    PluginManagerConfig
	db        *sql.DB
	mu        sync.RWMutex
	pluginDir string
}

// NewPluginService 创建插件管理服务
func NewPluginService(config PluginManagerConfig) *PluginService {
	return &PluginService{
		config:    config,
		pluginDir: config.PluginDir,
	}
}

// Initialize 初始化插件服务
func (ps *PluginService) Initialize() error {
	// 创建插件目录
	if err := os.MkdirAll(ps.pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// 初始化数据库
	dbPath := filepath.Join(ps.pluginDir, "plugins.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	ps.db = db

	// 创建表
	if err := ps.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// 扫描现有插件
	if err := ps.scanPlugins(); err != nil {
		log.Printf("Warning: failed to scan existing plugins: %v", err)
	}

	return nil
}

// createTables 创建数据库表
func (ps *PluginService) createTables() error {
	// 创建插件表
	pluginsQuery := `
	CREATE TABLE IF NOT EXISTS plugins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		path TEXT NOT NULL,
		size INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		modified_at TEXT NOT NULL,
		status TEXT NOT NULL,
		description TEXT,
		created_at_db TEXT DEFAULT CURRENT_TIMESTAMP
	)
	`

	// 创建过滤规则表
	filterRulesQuery := `
	CREATE TABLE IF NOT EXISTS filter_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		action TEXT NOT NULL,
		direction TEXT NOT NULL,
		source_ips TEXT, -- JSON格式存储
		dest_ips TEXT,   -- JSON格式存储
		source_ports TEXT, -- JSON格式存储
		dest_ports TEXT,   -- JSON格式存储
		enabled INTEGER NOT NULL DEFAULT 1,
		priority INTEGER NOT NULL DEFAULT 0,
		description TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		created_at_db TEXT DEFAULT CURRENT_TIMESTAMP
	)
	`

	// 创建过滤配置表
	filterConfigQuery := `
	CREATE TABLE IF NOT EXISTS filter_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		enabled INTEGER NOT NULL DEFAULT 0,
		default_action TEXT NOT NULL,
		rules TEXT, -- JSON格式存储完整规则列表
		update_time TEXT NOT NULL,
		created_at_db TEXT DEFAULT CURRENT_TIMESTAMP
	)
	`

	// 执行建表语句
	if _, err := ps.db.Exec(pluginsQuery); err != nil {
		return fmt.Errorf("failed to create plugins table: %w", err)
	}

	if _, err := ps.db.Exec(filterRulesQuery); err != nil {
		return fmt.Errorf("failed to create filter_rules table: %w", err)
	}

	if _, err := ps.db.Exec(filterConfigQuery); err != nil {
		return fmt.Errorf("failed to create filter_config table: %w", err)
	}

	// 插入默认过滤配置
	if err := ps.insertDefaultFilterConfig(); err != nil {
		return fmt.Errorf("failed to insert default filter config: %w", err)
	}

	return nil
}

// insertDefaultFilterConfig 插入默认过滤配置
func (ps *PluginService) insertDefaultFilterConfig() error {
	// 检查是否已存在配置
	var count int
	err := ps.db.QueryRow("SELECT COUNT(*) FROM filter_config").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 已存在配置，无需插入
	}

	// 插入默认配置
	query := `
		INSERT INTO filter_config (enabled, default_action, rules, update_time)
		VALUES (?, ?, ?, ?)
	`

	_, err = ps.db.Exec(query, 0, "allow", "[]", time.Now().Format(time.RFC3339))
	return err
}

// UploadPlugin 上传插件
func (ps *PluginService) UploadPlugin(name string, file io.Reader, metadata PluginUploadRequest) (*PluginInfo, error) {
	// 检查插件是否已存在
	exists, err := ps.pluginExists(name)
	if err != nil {
		return nil, fmt.Errorf("failed to check plugin existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("plugin '%s' already exists", name)
	}

	// 生成插件文件路径
	pluginPath := filepath.Join(ps.pluginDir, name)

	// 检查文件大小限制
	if ps.config.MaxSize > 0 {
		// 创建临时文件来检查大小
		tempFile, err := os.CreateTemp(ps.pluginDir, "upload-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tempFile.Name())

		// 写入文件并检查大小
		written, err := io.Copy(tempFile, file)
		if err != nil {
			tempFile.Close()
			return nil, fmt.Errorf("failed to write plugin file: %w", err)
		}

		if ps.config.MaxSize > 0 && written > ps.config.MaxSize {
			tempFile.Close()
			return nil, fmt.Errorf("plugin file too large: %d bytes (max: %d)", written, ps.config.MaxSize)
		}

		// 重命名到最终路径
		if err := tempFile.Close(); err != nil {
			return nil, fmt.Errorf("failed to close temp file: %w", err)
		}

		if err := os.Rename(tempFile.Name(), pluginPath); err != nil {
			return nil, fmt.Errorf("failed to move plugin file: %w", err)
		}
	} else {
		// 直接写入最终文件
		outFile, err := os.Create(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin file: %w", err)
		}
		defer outFile.Close()

		written, err := io.Copy(outFile, file)
		if err != nil {
			return nil, fmt.Errorf("failed to write plugin file: %w", err)
		}

		if ps.config.MaxSize > 0 && written > ps.config.MaxSize {
			return nil, fmt.Errorf("plugin file too large: %d bytes (max: %d)", written, ps.config.MaxSize)
		}
	}

	// 获取文件信息
	fileInfo, err := os.Stat(pluginPath)
	if err != nil {
		os.Remove(pluginPath)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	now := time.Now()
	// 插入数据库
	query := `
		INSERT INTO plugins (name, path, size, created_at, modified_at, status, description)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = ps.db.Exec(query, name, pluginPath, fileInfo.Size(),
		fileInfo.ModTime().Format(time.RFC3339), now.Format(time.RFC3339),
		"active", metadata.Description)
	if err != nil {
		os.Remove(pluginPath)
		return nil, fmt.Errorf("failed to insert plugin into database: %w", err)
	}

	// 创建插件信息
	pluginInfo := &PluginInfo{
		Name:        name,
		Path:        pluginPath,
		Size:        fileInfo.Size(),
		CreatedAt:   fileInfo.ModTime(),
		ModifiedAt:  now,
		Status:      "active",
		Description: metadata.Description,
	}

	log.Printf("Plugin '%s' uploaded successfully", name)
	return pluginInfo, nil
}

// DeletePlugin 删除插件
func (ps *PluginService) DeletePlugin(name string) error {
	// 获取插件信息
	plugin, err := ps.getPluginFromDB(name)
	if err != nil {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// 删除文件
	if err := os.Remove(plugin.Path); err != nil {
		log.Printf("Warning: failed to remove plugin file %s: %v", plugin.Path, err)
	}

	// 从数据库删除
	query := "DELETE FROM plugins WHERE name = ?"
	_, err = ps.db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete plugin from database: %w", err)
	}

	log.Printf("Plugin '%s' deleted successfully", name)
	return nil
}

// GetPlugin 获取插件信息
func (ps *PluginService) GetPlugin(name string) (*PluginInfo, error) {
	plugin, err := ps.getPluginFromDB(name)
	if err != nil {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	// 更新文件信息（大小等可能发生变化）
	if err := ps.updatePluginFileInfo(plugin); err != nil {
		log.Printf("Warning: failed to update plugin info for %s: %v", name, err)
	}

	return plugin, nil
}

// ListPlugins 获取插件列表
func (ps *PluginService) ListPlugins() ([]*PluginInfo, error) {
	query := "SELECT name, path, size, created_at, modified_at, status, description FROM plugins ORDER BY name"
	rows, err := ps.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*PluginInfo
	for rows.Next() {
		var name, path, status, description string
		var size int64
		var createdAtStr, modifiedAtStr string

		err := rows.Scan(&name, &path, &size, &createdAtStr, &modifiedAtStr, &status, &description)
		if err != nil {
			log.Printf("Warning: failed to scan plugin row: %v", err)
			continue
		}

		createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
		modifiedAt, _ := time.Parse(time.RFC3339, modifiedAtStr)

		plugin := &PluginInfo{
			Name:        name,
			Path:        path,
			Size:        size,
			CreatedAt:   createdAt,
			ModifiedAt:  modifiedAt,
			Status:      status,
			Description: description,
		}

		// 更新文件信息
		if err := ps.updatePluginFileInfo(plugin); err != nil {
			log.Printf("Warning: failed to update plugin info for %s: %v", name, err)
		}

		plugins = append(plugins, plugin)
	}

	// 按名称排序
	sort.Slice(plugins, func(i, j int) bool {
		return strings.ToLower(plugins[i].Name) < strings.ToLower(plugins[j].Name)
	})

	return plugins, nil
}

// scanPlugins 扫描插件目录
func (ps *PluginService) scanPlugins() error {
	return filepath.Walk(ps.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 跳过数据库文件
		if strings.HasSuffix(path, ".db") {
			return nil
		}

		// 只处理可执行文件或特定扩展名
		if !ps.isValidPluginFile(path, info) {
			return nil
		}

		name := filepath.Base(path)

		// 检查是否已在数据库中
		exists, err := ps.pluginExists(name)
		if err != nil {
			log.Printf("Warning: failed to check plugin existence for %s: %v", name, err)
			return nil
		}
		if exists {
			return nil
		}

		// 插入到数据库
		query := `
			INSERT INTO plugins (name, path, size, created_at, modified_at, status, description)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err = ps.db.Exec(query, name, path, info.Size(),
			info.ModTime().Format(time.RFC3339), info.ModTime().Format(time.RFC3339),
			"active", "")
		if err != nil {
			log.Printf("Warning: failed to insert plugin %s into database: %v", name, err)
		}

		return nil
	})
}

// pluginExists 检查插件是否存在
func (ps *PluginService) pluginExists(name string) (bool, error) {
	query := "SELECT COUNT(*) FROM plugins WHERE name = ?"
	var count int
	err := ps.db.QueryRow(query, name).Scan(&count)
	return count > 0, err
}

// getPluginFromDB 从数据库获取插件信息
func (ps *PluginService) getPluginFromDB(name string) (*PluginInfo, error) {
	query := "SELECT name, path, size, created_at, modified_at, status, description FROM plugins WHERE name = ?"
	var pluginName, path, status, description string
	var size int64
	var createdAtStr, modifiedAtStr string

	err := ps.db.QueryRow(query, name).Scan(&pluginName, &path, &size, &createdAtStr, &modifiedAtStr, &status, &description)
	if err != nil {
		return nil, err
	}

	createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
	modifiedAt, _ := time.Parse(time.RFC3339, modifiedAtStr)

	return &PluginInfo{
		Name:        pluginName,
		Path:        path,
		Size:        size,
		CreatedAt:   createdAt,
		ModifiedAt:  modifiedAt,
		Status:      status,
		Description: description,
	}, nil
}

// updatePluginFileInfo 更新插件文件信息
func (ps *PluginService) updatePluginFileInfo(plugin *PluginInfo) error {
	fileInfo, err := os.Stat(plugin.Path)
	if err != nil {
		return err
	}

	// 如果文件信息发生变化，更新数据库
	if fileInfo.Size() != plugin.Size || !fileInfo.ModTime().Equal(plugin.ModifiedAt) {
		plugin.Size = fileInfo.Size()
		plugin.ModifiedAt = fileInfo.ModTime()

		query := "UPDATE plugins SET size = ?, modified_at = ? WHERE name = ?"
		_, err := ps.db.Exec(query, plugin.Size, plugin.ModifiedAt.Format(time.RFC3339), plugin.Name)
		if err != nil {
			log.Printf("Warning: failed to update plugin info in database: %v", err)
		}
	}

	return nil
}

// validatePlugin 验证插件文件
func (ps *PluginService) validatePlugin(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	// 检查文件是否可执行（Unix系统）或检查扩展名（Windows）
	if ps.isValidPluginFile(path, fileInfo) {
		return nil
	}

	return fmt.Errorf("invalid plugin file: %s", path)
}

// isValidPluginFile 检查是否为有效的插件文件
func (ps *PluginService) isValidPluginFile(path string, info os.FileInfo) bool {
	// 跳过隐藏文件和目录
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}

	// 检查文件扩展名
	validExtensions := []string{".exe", ".so", ".dylib", ".bin", ".plugin"}
	ext := strings.ToLower(filepath.Ext(path))

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}

	// 检查文件是否可执行（Unix系统）
	if info.Mode()&0111 != 0 {
		return true
	}

	return false
}

// GetPluginDir 获取插件目录
func (ps *PluginService) GetPluginDir() string {
	return ps.pluginDir
}

// UpdatePluginMetadata 更新插件元数据
func (ps *PluginService) UpdatePluginMetadata(name string, metadata PluginUploadRequest) error {
	query := "UPDATE plugins SET description = ? WHERE name = ?"
	_, err := ps.db.Exec(query, metadata.Description, name)
	if err != nil {
		return fmt.Errorf("failed to update plugin metadata: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (ps *PluginService) Close() error {
	if ps.db != nil {
		return ps.db.Close()
	}
	return nil
}
