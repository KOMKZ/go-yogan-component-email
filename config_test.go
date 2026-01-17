package email

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Default: "mandrill",
				Drivers: map[string]map[string]any{
					"mandrill": {"api_key": "test"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing default",
			config: &Config{
				Drivers: map[string]map[string]any{
					"mandrill": {"api_key": "test"},
				},
			},
			wantErr: true,
		},
		{
			name: "no drivers",
			config: &Config{
				Default: "mandrill",
				Drivers: map[string]map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "default driver not configured",
			config: &Config{
				Default: "mandrill",
				Drivers: map[string]map[string]any{
					"smtp": {"host": "localhost"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ApplyDefaults(t *testing.T) {
	config := &Config{}
	config.ApplyDefaults()

	if config.Default != DriverMandrill {
		t.Errorf("expected default driver '%s', got '%s'", DriverMandrill, config.Default)
	}
}

func TestConfig_ApplyDefaults_NoOverwrite(t *testing.T) {
	config := &Config{
		Default: "custom",
	}
	config.ApplyDefaults()

	if config.Default != "custom" {
		t.Errorf("expected default driver 'custom', got '%s'", config.Default)
	}
}
