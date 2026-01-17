package email

import "context"

// Driver 邮件驱动接口
type Driver interface {
	// Name 驱动名称
	Name() string

	// Send 发送邮件（同步）
	Send(ctx context.Context, msg *Message) (*Result, error)

	// Validate 验证配置
	Validate() error
}

// DriverFactory 驱动工厂函数
type DriverFactory func(config map[string]any) (Driver, error)

// Registry 驱动注册表
type Registry struct {
	factories map[string]DriverFactory
}

// NewRegistry 创建驱动注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]DriverFactory),
	}
}

// Register 注册驱动工厂
func (r *Registry) Register(name string, factory DriverFactory) {
	r.factories[name] = factory
}

// Create 创建驱动实例
func (r *Registry) Create(name string, config map[string]any) (Driver, error) {
	factory, ok := r.factories[name]
	if !ok {
		return nil, ErrDriverNotFound.WithMsgf("驱动未找到: %s", name)
	}
	return factory(config)
}

// Has 检查驱动是否存在
func (r *Registry) Has(name string) bool {
	_, ok := r.factories[name]
	return ok
}

// Names 获取所有注册的驱动名称
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry 默认驱动注册表
var DefaultRegistry = NewRegistry()

// RegisterDriver 注册驱动到默认注册表
func RegisterDriver(name string, factory DriverFactory) {
	DefaultRegistry.Register(name, factory)
}
