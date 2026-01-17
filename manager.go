package email

import (
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
func NewManager(config *Config, registry *Registry, log *logger.CtxZapLogger) *Manager {
	if registry == nil {
		registry = DefaultRegistry
	}
	return &Manager{
		config:   config,
		registry: registry,
		drivers:  make(map[string]Driver),
		logger:   log,
	}
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

// Config 获取配置
func (m *Manager) Config() *Config {
	return m.config
}

// AvailableDrivers 获取可用驱动列表
func (m *Manager) AvailableDrivers() []string {
	return m.registry.Names()
}
