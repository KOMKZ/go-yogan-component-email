package email

import (
	"github.com/KOMKZ/go-yogan-framework/component"
	"github.com/KOMKZ/go-yogan-framework/logger"
)

const ComponentName = "email"

// Component 邮件组件
type Component struct {
	config   *Config
	manager  *Manager
	logger   *logger.CtxZapLogger
	registry component.Registry
}

// NewComponent 创建邮件组件
func NewComponent() *Component {
	return &Component{}
}

// Name 组件名称
func (c *Component) Name() string {
	return ComponentName
}

// Init 初始化组件
func (c *Component) Init(loader component.ConfigLoader) error {
	var cfg Config
	if err := loader.Unmarshal(ComponentName, &cfg); err != nil {
		// 配置不存在时使用默认配置
		cfg = Config{
			Default: DriverMandrill,
			Drivers: make(map[string]map[string]any),
		}
	}

	cfg.ApplyDefaults()

	// 如果没有配置任何驱动，不验证
	if len(cfg.Drivers) > 0 {
		if err := cfg.Validate(); err != nil {
			return ErrDriverConfig.Wrap(err)
		}
	}

	c.config = &cfg
	c.logger = logger.GetLogger(ComponentName)

	return nil
}

// Start 启动组件
func (c *Component) Start() error {
	c.manager = NewManager(c.config, DefaultRegistry, c.logger)
	return nil
}

// Stop 停止组件
func (c *Component) Stop() error {
	if c.manager != nil {
		return c.manager.Close()
	}
	return nil
}

// SetRegistry 设置组件注册表
func (c *Component) SetRegistry(registry any) {
	if r, ok := registry.(component.Registry); ok {
		c.registry = r
	}
}

// GetManager 获取邮件管理器
func (c *Component) GetManager() *Manager {
	return c.manager
}

// New 创建邮件构建器（快捷方法）
func (c *Component) New() *Builder {
	return c.manager.New()
}
