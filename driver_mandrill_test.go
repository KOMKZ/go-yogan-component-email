package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMandrillDriver(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]any{
				"api_key": "test-api-key",
			},
			wantErr: false,
		},
		{
			name:    "missing api key",
			config:  map[string]any{},
			wantErr: true,
		},
		{
			name: "custom base url",
			config: map[string]any{
				"api_key":  "test-api-key",
				"base_url": "https://custom.api.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := NewMandrillDriver(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMandrillDriver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && driver == nil {
				t.Error("expected driver, got nil")
			}
		})
	}
}

func TestMandrillDriver_Name(t *testing.T) {
	driver, _ := NewMandrillDriver(map[string]any{"api_key": "test"})
	if driver.Name() != DriverMandrill {
		t.Errorf("expected %s, got %s", DriverMandrill, driver.Name())
	}
}

func TestMandrillDriver_Send_Success(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/messages/send.json" {
			t.Errorf("expected /messages/send.json, got %s", r.URL.Path)
		}

		// 返回成功响应
		response := []map[string]any{
			{
				"_id":    "abc123",
				"email":  "user@example.com",
				"status": "sent",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, err := NewMandrillDriver(map[string]any{
		"api_key":  "test-api-key",
		"base_url": server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}

	msg := &Message{
		From:     "sender@example.com",
		FromName: "Sender",
		To:       []string{"user@example.com"},
		Subject:  "Test Email",
		BodyHTML: "<h1>Hello</h1>",
		BodyText: "Hello",
	}

	result, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MessageID != "abc123" {
		t.Errorf("expected MessageID 'abc123', got '%s'", result.MessageID)
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
}

func TestMandrillDriver_Send_WithCcBcc(t *testing.T) {
	var receivedPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)

		response := []map[string]any{
			{"_id": "test", "status": "sent"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:    "sender@example.com",
		To:      []string{"to@example.com"},
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"bcc@example.com"},
		Subject: "Test",
		BodyHTML: "Hello",
	}

	_, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证收件人列表
	message := receivedPayload["message"].(map[string]any)
	recipients := message["to"].([]any)

	if len(recipients) != 3 {
		t.Errorf("expected 3 recipients, got %d", len(recipients))
	}

	// 验证类型
	types := make(map[string]bool)
	for _, r := range recipients {
		recipient := r.(map[string]any)
		types[recipient["type"].(string)] = true
	}

	if !types["to"] || !types["cc"] || !types["bcc"] {
		t.Error("expected to, cc, bcc recipient types")
	}
}

func TestMandrillDriver_Send_WithAttachments(t *testing.T) {
	var receivedPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)

		response := []map[string]any{
			{"_id": "test", "status": "sent"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
		Attachments: []Attachment{
			{Filename: "file.pdf", Content: []byte("pdf content")},
			{Filename: "logo.png", Content: []byte("logo"), Inline: true, ContentID: "logo"},
		},
	}

	_, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	message := receivedPayload["message"].(map[string]any)

	// 验证普通附件
	attachments := message["attachments"].([]any)
	if len(attachments) != 1 {
		t.Errorf("expected 1 attachment, got %d", len(attachments))
	}

	// 验证内联图片
	images := message["images"].([]any)
	if len(images) != 1 {
		t.Errorf("expected 1 image, got %d", len(images))
	}
}

func TestMandrillDriver_Send_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"status":  "error",
			"code":    -1,
			"name":    "Invalid_Key",
			"message": "Invalid API key",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "invalid-key",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
	}

	_, err := driver.Send(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid API key")
	}
}

func TestMandrillDriver_Send_Rejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []map[string]any{
			{
				"_id":           "test",
				"email":         "user@example.com",
				"status":        "rejected",
				"reject_reason": "hard-bounce",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
	}

	result, err := driver.Send(context.Background(), msg)
	if err == nil {
		t.Error("expected error for rejected email")
	}
	if result == nil {
		t.Error("expected result even on rejection")
	}
	if result != nil && result.Success {
		t.Error("expected Success to be false")
	}
}

func TestMandrillDriver_Send_InvalidMessage(t *testing.T) {
	driver, _ := NewMandrillDriver(map[string]any{
		"api_key": "test",
	})

	// 无收件人
	msg := &Message{
		From:     "sender@example.com",
		Subject:  "Test",
		BodyHTML: "Hello",
	}

	_, err := driver.Send(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid message")
	}
}

func TestMandrillDriver_Send_WithReplyTo(t *testing.T) {
	var receivedPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)

		response := []map[string]any{
			{"_id": "test", "status": "sent"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
		ReplyTo:  "reply@example.com",
	}

	_, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证 Reply-To 头
	message := receivedPayload["message"].(map[string]any)
	headers, ok := message["headers"].(map[string]any)
	if !ok {
		t.Error("expected headers in message")
	}
	if headers["Reply-To"] != "reply@example.com" {
		t.Errorf("expected Reply-To header, got %v", headers["Reply-To"])
	}
}

func TestMandrillDriver_Send_WithCustomHeaders(t *testing.T) {
	var receivedPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)

		response := []map[string]any{
			{"_id": "test", "status": "sent"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
		Headers: map[string]string{
			"X-Custom": "value",
		},
	}

	_, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证自定义头
	message := receivedPayload["message"].(map[string]any)
	headers, ok := message["headers"].(map[string]any)
	if !ok {
		t.Error("expected headers in message")
	}
	if headers["X-Custom"] != "value" {
		t.Errorf("expected X-Custom header, got %v", headers["X-Custom"])
	}
}

func TestMandrillDriver_Send_TextOnly(t *testing.T) {
	var receivedPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)

		response := []map[string]any{
			{"_id": "test", "status": "sent"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyText: "Hello plain text",
	}

	_, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	message := receivedPayload["message"].(map[string]any)
	if message["text"] != "Hello plain text" {
		t.Errorf("expected text body, got %v", message["text"])
	}
}

func TestMandrillDriver_Send_QueuedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []map[string]any{
			{"_id": "test", "status": "queued"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
	}

	result, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success for queued status")
	}
}

func TestMandrillDriver_Send_ScheduledStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []map[string]any{
			{"_id": "test", "status": "scheduled"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
	}

	result, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success for scheduled status")
	}
}

func TestMandrillDriver_Send_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
	}

	_, err := driver.Send(context.Background(), msg)
	if err == nil {
		t.Error("expected error for empty response")
	}
}

func TestMandrillDriver_Send_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	driver, _ := NewMandrillDriver(map[string]any{
		"api_key":  "test",
		"base_url": server.URL,
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyHTML: "Hello",
	}

	_, err := driver.Send(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
