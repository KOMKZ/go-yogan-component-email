package email

import (
	"context"
	"testing"

	"github.com/KOMKZ/go-yogan-framework/logger"
)

func TestBuilder_ChainedCalls(t *testing.T) {
	log := logger.GetLogger("test")

	// 创建模拟驱动
	mockDriver := &MockDriver{
		name: "mock",
		sendResult: &Result{
			MessageID: "test-123",
			Status:    "sent",
			Success:   true,
		},
	}

	// 创建注册表
	registry := NewRegistry()
	registry.Register("mock", func(config map[string]any) (Driver, error) {
		return mockDriver, nil
	})

	// 创建配置
	config := &Config{
		Default:         "mock",
		DefaultFrom:     "default@example.com",
		DefaultFromName: "Default Sender",
		Drivers: map[string]map[string]any{
			"mock": {},
		},
	}

	// 创建管理器
	manager, err := NewManager(config, log, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 测试链式调用
	result, err := manager.New().
		To("user@example.com").
		Cc("cc@example.com").
		Bcc("bcc@example.com").
		Subject("Test Subject").
		Body("<h1>Hello</h1>").
		BodyText("Hello").
		ReplyTo("reply@example.com").
		Header("X-Custom", "value").
		Send(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MessageID != "test-123" {
		t.Errorf("expected message ID 'test-123', got '%s'", result.MessageID)
	}
}

func TestBuilder_DefaultFrom(t *testing.T) {
	log := logger.GetLogger("test")

	mockDriver := &MockDriver{
		name:       "mock",
		sendResult: &Result{Success: true},
	}

	registry := NewRegistry()
	registry.Register("mock", func(config map[string]any) (Driver, error) {
		return mockDriver, nil
	})

	config := &Config{
		Default:         "mock",
		DefaultFrom:     "default@example.com",
		DefaultFromName: "Default Name",
		Drivers: map[string]map[string]any{
			"mock": {},
		},
	}

	manager, err := NewManager(config, log, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := manager.New().
		To("user@example.com").
		Subject("Test").
		Body("Hello")

	// 验证默认值应用
	msg := builder.Message()
	if msg.From != "" {
		t.Error("From should be empty before send")
	}

	// 发送后检查
	_, _ = builder.Send(context.Background())
	if msg.From != "default@example.com" {
		t.Errorf("expected default from, got '%s'", msg.From)
	}
}

func TestBuilder_OverrideFrom(t *testing.T) {
	log := logger.GetLogger("test")

	mockDriver := &MockDriver{
		name:       "mock",
		sendResult: &Result{Success: true},
	}

	registry := NewRegistry()
	registry.Register("mock", func(config map[string]any) (Driver, error) {
		return mockDriver, nil
	})

	config := &Config{
		Default:     "mock",
		DefaultFrom: "default@example.com",
		Drivers: map[string]map[string]any{
			"mock": {},
		},
	}

	manager, err := NewManager(config, log, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builder := manager.New().
		From("custom@example.com").
		To("user@example.com").
		Subject("Test").
		Body("Hello")

	_, _ = builder.Send(context.Background())

	msg := builder.Message()
	if msg.From != "custom@example.com" {
		t.Errorf("expected custom from, got '%s'", msg.From)
	}
}

func TestBuilder_Attachments(t *testing.T) {
	log := logger.GetLogger("test")

	mockDriver := &MockDriver{
		name:       "mock",
		sendResult: &Result{Success: true},
	}

	registry := NewRegistry()
	registry.Register("mock", func(config map[string]any) (Driver, error) {
		return mockDriver, nil
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

	builder := manager.New().
		To("user@example.com").
		Subject("Test").
		Body("Hello").
		Attach("file.pdf", []byte("pdf content")).
		AttachWithType("image.jpg", []byte("image"), "image/jpeg").
		Embed("logo", "logo.png", []byte("logo"))

	msg := builder.Message()
	if len(msg.Attachments) != 3 {
		t.Errorf("expected 3 attachments, got %d", len(msg.Attachments))
	}

	// 验证内联附件
	found := false
	for _, att := range msg.Attachments {
		if att.Inline && att.ContentID == "logo" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected inline attachment with ContentID 'logo'")
	}
}

func TestBuilder_SwitchDriver(t *testing.T) {
	log := logger.GetLogger("test")

	mockDriver1 := &MockDriver{
		name:       "driver1",
		sendResult: &Result{MessageID: "from-driver1", Success: true},
	}
	mockDriver2 := &MockDriver{
		name:       "driver2",
		sendResult: &Result{MessageID: "from-driver2", Success: true},
	}

	registry := NewRegistry()
	registry.Register("driver1", func(config map[string]any) (Driver, error) {
		return mockDriver1, nil
	})
	registry.Register("driver2", func(config map[string]any) (Driver, error) {
		return mockDriver2, nil
	})

	config := &Config{
		Default: "driver1",
		Drivers: map[string]map[string]any{
			"driver1": {},
			"driver2": {},
		},
	}

	manager, err := NewManager(config, log, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 使用默认驱动
	result1, _ := manager.New().
		To("user@example.com").
		Subject("Test").
		Body("Hello").
		Send(context.Background())

	if result1.MessageID != "from-driver1" {
		t.Errorf("expected from-driver1, got %s", result1.MessageID)
	}

	// 切换驱动
	result2, _ := manager.New().
		Driver("driver2").
		To("user@example.com").
		Subject("Test").
		Body("Hello").
		Send(context.Background())

	if result2.MessageID != "from-driver2" {
		t.Errorf("expected from-driver2, got %s", result2.MessageID)
	}
}
