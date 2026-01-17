package email

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	config := &Config{
		Default: "mock",
		Drivers: map[string]map[string]any{
			"mock": {},
		},
	}

	// 使用自定义注册表
	registry := NewRegistry()
	registry.Register("mock", func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "mock"}, nil
	})

	manager := NewManager(config, registry, nil)
	if manager == nil {
		t.Error("expected manager, got nil")
	}
}

func TestNewManager_NilRegistry(t *testing.T) {
	config := &Config{
		Default: DriverMandrill,
		Drivers: map[string]map[string]any{
			DriverMandrill: {"api_key": "test"},
		},
	}

	// 使用默认注册表
	manager := NewManager(config, nil, nil)
	if manager == nil {
		t.Error("expected manager, got nil")
	}
}

func TestManager_GetDriver(t *testing.T) {
	registry := NewRegistry()
	registry.Register("mock", func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "mock"}, nil
	})

	config := &Config{
		Default: "mock",
		Drivers: map[string]map[string]any{
			"mock": {},
		},
	}

	manager := NewManager(config, registry, nil)

	// 第一次获取
	driver1, err := manager.GetDriver("mock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver1 == nil {
		t.Error("expected driver, got nil")
	}

	// 第二次获取（应该返回缓存的驱动）
	driver2, err := manager.GetDriver("mock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver1 != driver2 {
		t.Error("expected same driver instance")
	}
}

func TestManager_GetDriver_NotConfigured(t *testing.T) {
	registry := NewRegistry()
	config := &Config{
		Default: "mock",
		Drivers: map[string]map[string]any{},
	}

	manager := NewManager(config, registry, nil)

	_, err := manager.GetDriver("nonexistent")
	if err == nil {
		t.Error("expected error for unconfigured driver")
	}
}

func TestManager_New(t *testing.T) {
	registry := NewRegistry()
	registry.Register("mock", func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "mock"}, nil
	})

	config := &Config{
		Default:         "mock",
		DefaultFrom:     "default@example.com",
		DefaultFromName: "Default",
		Drivers: map[string]map[string]any{
			"mock": {},
		},
	}

	manager := NewManager(config, registry, nil)

	builder := manager.New()
	if builder == nil {
		t.Error("expected builder, got nil")
	}
	if builder.driver != "mock" {
		t.Errorf("expected driver 'mock', got '%s'", builder.driver)
	}
}

func TestManager_Close(t *testing.T) {
	registry := NewRegistry()
	registry.Register("mock", func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "mock"}, nil
	})

	config := &Config{
		Default: "mock",
		Drivers: map[string]map[string]any{
			"mock": {},
		},
	}

	manager := NewManager(config, registry, nil)

	// 创建驱动
	_, _ = manager.GetDriver("mock")

	// 关闭
	if err := manager.Close(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 关闭后应该重新创建驱动
	driver, err := manager.GetDriver("mock")
	if err != nil {
		t.Fatalf("unexpected error after close: %v", err)
	}
	if driver == nil {
		t.Error("expected driver after close")
	}
}

func TestManager_Config(t *testing.T) {
	config := &Config{
		Default: "mock",
	}

	manager := NewManager(config, nil, nil)
	if manager.Config() != config {
		t.Error("expected same config")
	}
}

func TestManager_AvailableDrivers(t *testing.T) {
	registry := NewRegistry()
	registry.Register("driver1", func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "driver1"}, nil
	})
	registry.Register("driver2", func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "driver2"}, nil
	})

	config := &Config{
		Default: "driver1",
	}

	manager := NewManager(config, registry, nil)
	drivers := manager.AvailableDrivers()

	if len(drivers) != 2 {
		t.Errorf("expected 2 drivers, got %d", len(drivers))
	}
}
