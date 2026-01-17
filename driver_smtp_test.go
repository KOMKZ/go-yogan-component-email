package email

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNewSMTPDriver(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]any{
				"host":     "smtp.example.com",
				"port":     587,
				"username": "user",
				"password": "pass",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: map[string]any{
				"port": 587,
			},
			wantErr: true,
		},
		{
			name: "with float port",
			config: map[string]any{
				"host": "smtp.example.com",
				"port": float64(587),
			},
			wantErr: false,
		},
		{
			name: "with security option",
			config: map[string]any{
				"host":     "smtp.example.com",
				"security": "starttls",
			},
			wantErr: false,
		},
		{
			name: "with timeout",
			config: map[string]any{
				"host":    "smtp.example.com",
				"timeout": "60s",
			},
			wantErr: false,
		},
		{
			name: "with local name",
			config: map[string]any{
				"host":       "smtp.example.com",
				"local_name": "localhost",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := NewSMTPDriver(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSMTPDriver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && driver == nil {
				t.Error("expected driver, got nil")
			}
		})
	}
}

func TestSMTPDriver_Name(t *testing.T) {
	driver, _ := NewSMTPDriver(map[string]any{"host": "smtp.example.com"})
	if driver.Name() != DriverSMTP {
		t.Errorf("expected %s, got %s", DriverSMTP, driver.Name())
	}
}

func TestSMTPDriver_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *SMTPConfig
		wantErr bool
	}{
		{
			name: "valid",
			config: &SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: &SMTPConfig{
				Port: 587,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &SMTPConfig{
				Host: "smtp.example.com",
				Port: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &SMTPDriver{config: tt.config}
			err := driver.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSMTPDriver_buildEmailBody(t *testing.T) {
	driver := &SMTPDriver{
		config: &SMTPConfig{
			Host: "smtp.example.com",
			Port: 587,
		},
	}

	tests := []struct {
		name     string
		msg      *Message
		contains []string
	}{
		{
			name: "basic message",
			msg: &Message{
				From:     "sender@example.com",
				To:       []string{"to@example.com"},
				Subject:  "Test",
				BodyHTML: "<h1>Hello</h1>",
			},
			contains: []string{
				"From: sender@example.com",
				"To: to@example.com",
				"Subject:",
				"text/html",
			},
		},
		{
			name: "with from name",
			msg: &Message{
				From:     "sender@example.com",
				FromName: "Sender Name",
				To:       []string{"to@example.com"},
				Subject:  "Test",
				BodyText: "Hello",
			},
			contains: []string{
				"Sender Name",
				"sender@example.com",
			},
		},
		{
			name: "with cc",
			msg: &Message{
				From:     "sender@example.com",
				To:       []string{"to@example.com"},
				Cc:       []string{"cc@example.com"},
				Subject:  "Test",
				BodyText: "Hello",
			},
			contains: []string{
				"Cc: cc@example.com",
			},
		},
		{
			name: "with reply-to",
			msg: &Message{
				From:     "sender@example.com",
				To:       []string{"to@example.com"},
				ReplyTo:  "reply@example.com",
				Subject:  "Test",
				BodyText: "Hello",
			},
			contains: []string{
				"Reply-To: reply@example.com",
			},
		},
		{
			name: "with custom headers",
			msg: &Message{
				From:    "sender@example.com",
				To:      []string{"to@example.com"},
				Subject: "Test",
				BodyText: "Hello",
				Headers: map[string]string{
					"X-Custom": "value",
				},
			},
			contains: []string{
				"X-Custom: value",
			},
		},
		{
			name: "with both html and text",
			msg: &Message{
				From:     "sender@example.com",
				To:       []string{"to@example.com"},
				Subject:  "Test",
				BodyHTML: "<h1>Hello</h1>",
				BodyText: "Hello",
			},
			contains: []string{
				"multipart/alternative",
				"text/plain",
				"text/html",
			},
		},
		{
			name: "with attachment",
			msg: &Message{
				From:     "sender@example.com",
				To:       []string{"to@example.com"},
				Subject:  "Test",
				BodyText: "Hello",
				Attachments: []Attachment{
					{Filename: "file.pdf", Content: []byte("pdf content")},
				},
			},
			contains: []string{
				"multipart/mixed",
				"file.pdf",
				"attachment",
			},
		},
		{
			name: "with inline attachment",
			msg: &Message{
				From:     "sender@example.com",
				To:       []string{"to@example.com"},
				Subject:  "Test",
				BodyHTML: "<img src='cid:logo'>",
				Attachments: []Attachment{
					{Filename: "logo.png", Content: []byte("logo"), Inline: true, ContentID: "logo"},
				},
			},
			contains: []string{
				"Content-ID: <logo>",
				"inline",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := driver.buildEmailBody(tt.msg)
			for _, s := range tt.contains {
				if !strings.Contains(body, s) {
					t.Errorf("expected body to contain %q, got:\n%s", s, body)
				}
			}
		})
	}
}

func TestSMTPDriver_Send_InvalidMessage(t *testing.T) {
	driver, _ := NewSMTPDriver(map[string]any{
		"host": "smtp.example.com",
		"port": 587,
	})

	msg := &Message{
		From:     "sender@example.com",
		Subject:  "Test",
		BodyText: "Hello",
		// Missing To
	}

	_, err := driver.Send(context.Background(), msg)
	if err == nil {
		t.Error("expected error for invalid message")
	}
}

func TestSMTPDriver_Send_ConnectionFailed(t *testing.T) {
	driver, _ := NewSMTPDriver(map[string]any{
		"host":    "nonexistent.invalid",
		"port":    12345,
		"timeout": "1s",
	})

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyText: "Hello",
	}

	_, err := driver.Send(context.Background(), msg)
	if err == nil {
		t.Error("expected connection error")
	}
}

// mockSMTPServer 创建一个简单的模拟 SMTP 服务器用于测试
func startMockSMTPServer(t *testing.T, responses []string) (string, func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}

	done := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// 发送欢迎消息
		conn.Write([]byte("220 Mock SMTP Server\r\n"))

		buf := make([]byte, 1024)
		for {
			select {
			case <-done:
				return
			default:
			}

			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := conn.Read(buf)
			if err != nil {
				continue
			}

			cmd := strings.ToUpper(strings.TrimSpace(string(buf[:n])))

			if strings.HasPrefix(cmd, "EHLO") {
				conn.Write([]byte("250-localhost\r\n250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "MAIL FROM") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "RCPT TO") {
				conn.Write([]byte("250 OK\r\n"))
			} else if cmd == "DATA" {
				conn.Write([]byte("354 Start mail input\r\n"))
			} else if strings.HasSuffix(cmd, ".") {
				conn.Write([]byte("250 OK\r\n"))
			} else if cmd == "QUIT" {
				conn.Write([]byte("221 Bye\r\n"))
				return
			}
		}
	}()

	cleanup := func() {
		close(done)
		listener.Close()
	}

	return listener.Addr().String(), cleanup
}

func TestSMTPDriver_Send_Success_NoAuth(t *testing.T) {
	addr, cleanup := startMockSMTPServer(t, nil)
	defer cleanup()

	parts := strings.Split(addr, ":")
	host := parts[0]
	port := 0
	if len(parts) > 1 {
		var err error
		port, err = parsePort(parts[1])
		if err != nil {
			t.Fatalf("failed to parse port: %v", err)
		}
	}

	driver, err := NewSMTPDriver(map[string]any{
		"host":     host,
		"port":     port,
		"security": "none",
		"timeout":  "5s",
	})
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}

	msg := &Message{
		From:     "sender@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyText: "Hello",
	}

	result, err := driver.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
	if result.Status != "sent" {
		t.Errorf("expected status 'sent', got '%s'", result.Status)
	}
}

func parsePort(s string) (int, error) {
	var port int
	_, err := strings.NewReader(s).Read(make([]byte, 0))
	if err != nil {
		return 0, err
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		port = port*10 + int(c-'0')
	}
	return port, nil
}

func TestDefaultRegistry_SMTP(t *testing.T) {
	if !DefaultRegistry.Has(DriverSMTP) {
		t.Error("expected DefaultRegistry to have SMTP driver")
	}
}
