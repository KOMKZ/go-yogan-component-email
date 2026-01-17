package email

import (
	"context"
	"testing"
)

// MockDriver 模拟驱动
type MockDriver struct {
	name       string
	sendResult *Result
	sendErr    error
}

func (d *MockDriver) Name() string {
	return d.name
}

func (d *MockDriver) Send(ctx context.Context, msg *Message) (*Result, error) {
	if d.sendErr != nil {
		return nil, d.sendErr
	}
	return d.sendResult, nil
}

func (d *MockDriver) Validate() error {
	return nil
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	factory := func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "test"}, nil
	}

	registry.Register("test", factory)

	if !registry.Has("test") {
		t.Error("expected registry to have 'test' driver")
	}
}

func TestRegistry_Create(t *testing.T) {
	registry := NewRegistry()

	factory := func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "test"}, nil
	}
	registry.Register("test", factory)

	driver, err := registry.Create("test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver.Name() != "test" {
		t.Errorf("expected driver name 'test', got '%s'", driver.Name())
	}
}

func TestRegistry_Create_NotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Create("nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent driver")
	}
}

func TestRegistry_Names(t *testing.T) {
	registry := NewRegistry()

	registry.Register("driver1", func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "driver1"}, nil
	})
	registry.Register("driver2", func(config map[string]any) (Driver, error) {
		return &MockDriver{name: "driver2"}, nil
	})

	names := registry.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 drivers, got %d", len(names))
	}
}

func TestDefaultRegistry(t *testing.T) {
	// 验证默认注册表已注册 Mandrill 驱动
	if !DefaultRegistry.Has(DriverMandrill) {
		t.Error("expected DefaultRegistry to have Mandrill driver")
	}
}
