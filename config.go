package email

import "fmt"

// ComponentName 组件名称（用于配置 key）
const ComponentName = "email"

// Config 邮件组件配置
type Config struct {
	// Default 默认驱动名称
	Default string `mapstructure:"default"`

	// DefaultFrom 默认发件人地址
	DefaultFrom string `mapstructure:"default_from"`

	// DefaultFromName 默认发件人名称
	DefaultFromName string `mapstructure:"default_from_name"`

	// Drivers 驱动配置
	Drivers map[string]map[string]any `mapstructure:"drivers"`
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Default == "" {
		return fmt.Errorf("default driver is required")
	}
	if len(c.Drivers) == 0 {
		return fmt.Errorf("at least one driver must be configured")
	}
	if _, ok := c.Drivers[c.Default]; !ok {
		return fmt.Errorf("default driver '%s' is not configured", c.Default)
	}
	return nil
}

// ApplyDefaults 应用默认值
func (c *Config) ApplyDefaults() {
	if c.Default == "" {
		c.Default = DriverMandrill
	}
}
