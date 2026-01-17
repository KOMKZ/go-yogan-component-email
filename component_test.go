package email

import (
	"testing"
)

// MockConfigLoader 模拟配置加载器
type MockConfigLoader struct {
	data   map[string]any
	hasKey bool
}

func (m *MockConfigLoader) Get(key string) interface{} {
	if m.data != nil {
		return m.data[key]
	}
	return nil
}

func (m *MockConfigLoader) Unmarshal(key string, out any) error {
	if !m.hasKey {
		return nil // 模拟配置不存在
	}
	return nil
}

func (m *MockConfigLoader) GetString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *MockConfigLoader) GetInt(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *MockConfigLoader) GetBool(key string) bool {
	if v, ok := m.data[key].(bool); ok {
		return v
	}
	return false
}

func (m *MockConfigLoader) IsSet(key string) bool {
	if m.data == nil {
		return false
	}
	_, ok := m.data[key]
	return ok
}

func TestNewComponent(t *testing.T) {
	comp := NewComponent()
	if comp == nil {
		t.Error("expected component, got nil")
	}
}

func TestComponent_Name(t *testing.T) {
	comp := NewComponent()
	if comp.Name() != ComponentName {
		t.Errorf("expected name '%s', got '%s'", ComponentName, comp.Name())
	}
}

func TestComponent_Init_NoConfig(t *testing.T) {
	comp := NewComponent()
	loader := &MockConfigLoader{hasKey: false}

	err := comp.Init(loader)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 应该使用默认配置
	if comp.config == nil {
		t.Error("expected config, got nil")
	}
	if comp.config.Default != DriverMandrill {
		t.Errorf("expected default driver '%s', got '%s'", DriverMandrill, comp.config.Default)
	}
}

func TestComponent_Start(t *testing.T) {
	comp := NewComponent()
	loader := &MockConfigLoader{hasKey: false}

	if err := comp.Init(loader); err != nil {
		t.Fatalf("init error: %v", err)
	}

	if err := comp.Start(); err != nil {
		t.Errorf("start error: %v", err)
	}

	if comp.GetManager() == nil {
		t.Error("expected manager after start")
	}
}

func TestComponent_Stop(t *testing.T) {
	comp := NewComponent()
	loader := &MockConfigLoader{hasKey: false}

	if err := comp.Init(loader); err != nil {
		t.Fatalf("init error: %v", err)
	}
	if err := comp.Start(); err != nil {
		t.Fatalf("start error: %v", err)
	}
	if err := comp.Stop(); err != nil {
		t.Errorf("stop error: %v", err)
	}
}

func TestComponent_Stop_NotStarted(t *testing.T) {
	comp := NewComponent()

	// 未启动时 stop 应该安全
	if err := comp.Stop(); err != nil {
		t.Errorf("stop error: %v", err)
	}
}

func TestComponent_SetRegistry(t *testing.T) {
	comp := NewComponent()

	// 传入非 Registry 类型应该被忽略
	comp.SetRegistry("not a registry")

	if comp.registry != nil {
		t.Error("expected nil registry for invalid type")
	}
}

func TestComponent_New(t *testing.T) {
	comp := NewComponent()
	loader := &MockConfigLoader{hasKey: false}

	if err := comp.Init(loader); err != nil {
		t.Fatalf("init error: %v", err)
	}
	if err := comp.Start(); err != nil {
		t.Fatalf("start error: %v", err)
	}

	builder := comp.New()
	if builder == nil {
		t.Error("expected builder, got nil")
	}
}
