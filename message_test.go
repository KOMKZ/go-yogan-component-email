package email

import (
	"testing"
)

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     *Message
		wantErr bool
	}{
		{
			name: "valid message with HTML",
			msg: &Message{
				To:       []string{"user@example.com"},
				Subject:  "Test",
				BodyHTML: "<p>Hello</p>",
			},
			wantErr: false,
		},
		{
			name: "valid message with text",
			msg: &Message{
				To:       []string{"user@example.com"},
				Subject:  "Test",
				BodyText: "Hello",
			},
			wantErr: false,
		},
		{
			name: "missing recipient",
			msg: &Message{
				Subject:  "Test",
				BodyHTML: "<p>Hello</p>",
			},
			wantErr: true,
		},
		{
			name: "missing subject",
			msg: &Message{
				To:       []string{"user@example.com"},
				BodyHTML: "<p>Hello</p>",
			},
			wantErr: true,
		},
		{
			name: "missing body",
			msg: &Message{
				To:      []string{"user@example.com"},
				Subject: "Test",
			},
			wantErr: true,
		},
		{
			name: "empty recipient list",
			msg: &Message{
				To:       []string{},
				Subject:  "Test",
				BodyHTML: "<p>Hello</p>",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
