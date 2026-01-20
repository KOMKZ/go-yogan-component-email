package email

import (
	"testing"

	"github.com/KOMKZ/go-yogan-framework/logger"
)

func TestManager_GetDriver(t *testing.T) {
	log := logger.GetLogger("test")
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

	manager, err := NewManager(config, log, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	log := logger.GetLogger("test")
	config := &Config{
		Default: "mock",
		Drivers: map[string]map[string]any{},
	}

	manager, err := NewManager(config, log, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = manager.GetDriver("nonexistent")
	if err == nil {
		t.Error("expected error for unconfigured driver")
	}
}

func TestManager_New(t *testing.T) {
	log := logger.GetLogger("test")
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

	manager, err := NewManager(config, log, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := manager.New()
	if builder == nil {
		t.Error("expected builder, got nil")
	}
	if builder.driver != "mock" {
		t.Errorf("expected driver 'mock', got '%s'", builder.driver)
	}
}

func TestManager_Close(t *testing.T) {
	log := logger.GetLogger("test")
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

	manager, err := NewManager(config, log, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	log := logger.GetLogger("test")
	config := &Config{
		Default: "mock",
	}

	manager, err := NewManager(config, log, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager.Config() != config {
		t.Error("expected same config")
	}
}

func TestManager_AvailableDrivers(t *testing.T) {
	log := logger.GetLogger("test")
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

	manager, err := NewManager(config, log, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	drivers := manager.AvailableDrivers()
	if len(drivers) != 2 {
		t.Errorf("expected 2 drivers, got %d", len(drivers))
	}
}
