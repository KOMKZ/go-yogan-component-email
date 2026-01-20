package email

import (
	"fmt"
	"sync"

	"github.com/KOMKZ/go-yogan-framework/logger"
	"go.uber.org/zap"
)

// Manager 邮件管理器
type Manager struct {
	config   *Config
	registry *Registry
	drivers  map[string]Driver
	logger   *logger.CtxZapLogger
	mu       sync.RWMutex
}

// NewManager 创建邮件管理器
// config: 邮件配置（必需）
// logger: 业务日志器（必需）
// registry: 驱动注册表（可选，为 nil 时使用 DefaultRegistry）
func NewManager(config *Config, log *logger.CtxZapLogger, registry *Registry) (*Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("email config cannot be nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if registry == nil {
		registry = DefaultRegistry
	}

	// 应用默认值并验证
	config.ApplyDefaults()
	if len(config.Drivers) > 0 {
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid email config: %w", err)
		}
	}

	return &Manager{
		config:   config,
		registry: registry,
		drivers:  make(map[string]Driver),
		logger:   log,
	}, nil
}

// GetDriver 获取驱动实例
func (m *Manager) GetDriver(name string) (Driver, error) {
	m.mu.RLock()
	driver, ok := m.drivers[name]
	m.mu.RUnlock()

	if ok {
		return driver, nil
	}

	// 创建驱动
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if driver, ok := m.drivers[name]; ok {
		return driver, nil
	}

	// 获取驱动配置
	driverConfig, ok := m.config.Drivers[name]
	if !ok {
		return nil, ErrDriverNotFound.WithMsgf("驱动配置未找到: %s", name)
	}

	// 创建驱动实例
	driver, err := m.registry.Create(name, driverConfig)
	if err != nil {
		return nil, err
	}

	m.drivers[name] = driver

	if m.logger != nil {
		m.logger.Info("email driver created", zap.String("driver", name))
	}

	return driver, nil
}

// New 创建邮件构建器
func (m *Manager) New() *Builder {
	return &Builder{
		manager: m,
		driver:  m.config.Default,
		message: &Message{},
	}
}

// Close 关闭管理器
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.drivers = make(map[string]Driver)
	return nil
}

// Shutdown 实现 samber/do.Shutdownable 接口
// 用于在 DI 容器关闭时自动关闭资源
func (m *Manager) Shutdown() error {
	return m.Close()
}

// Config 获取配置
func (m *Manager) Config() *Config {
	return m.config
}

// AvailableDrivers 获取可用驱动列表
func (m *Manager) AvailableDrivers() []string {
	return m.registry.Names()
}
