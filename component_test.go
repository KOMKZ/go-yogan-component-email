package email

import (
	"context"
	"testing"

	"github.com/KOMKZ/go-yogan-framework/logger"
)

func TestNewManager(t *testing.T) {
	log := logger.GetLogger("test")
	config := &Config{
		Default: DriverMandrill,
		Drivers: make(map[string]map[string]any),
	}

	manager, err := NewManager(config, log, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manager == nil {
		t.Fatal("expected manager, got nil")
	}
}

func TestNewManager_NilConfig(t *testing.T) {
	log := logger.GetLogger("test")

	_, err := NewManager(nil, log, nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestNewManager_NilLogger(t *testing.T) {
	config := &Config{
		Default: DriverMandrill,
		Drivers: make(map[string]map[string]any),
	}

	_, err := NewManager(config, nil, nil)
	if err == nil {
		t.Error("expected error for nil logger")
	}
}

func TestManager_NewBuilder(t *testing.T) {
	log := logger.GetLogger("test")
	config := &Config{
		Default: DriverMandrill,
		Drivers: make(map[string]map[string]any),
	}

	manager, err := NewManager(config, log, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := manager.New()
	if builder == nil {
		t.Error("expected builder, got nil")
	}
}

func TestManager_Shutdown(t *testing.T) {
	log := logger.GetLogger("test")
	config := &Config{
		Default: DriverMandrill,
		Drivers: make(map[string]map[string]any),
	}

	manager, err := NewManager(config, log, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = manager.Shutdown()
	if err != nil {
		t.Errorf("shutdown error: %v", err)
	}
}

func TestManager_GetDriver_NotFound(t *testing.T) {
	log := logger.GetLogger("test")
	config := &Config{
		Default: DriverMandrill,
		Drivers: make(map[string]map[string]any),
	}

	manager, err := NewManager(config, log, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = manager.GetDriver("non-existent")
	if err == nil {
		t.Error("expected error for non-existent driver")
	}
}

func TestBuilder_Send_NoDriver(t *testing.T) {
	log := logger.GetLogger("test")
	config := &Config{
		Default: DriverMandrill,
		Drivers: make(map[string]map[string]any),
	}

	manager, err := NewManager(config, log, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := manager.New()
	_, err = builder.
		To("test@example.com").
		Subject("Test").
		Body("<p>Test</p>").
		Send(context.Background())

	// 应该报错因为没有配置驱动
	if err == nil {
		t.Error("expected error when no driver configured")
	}
}
