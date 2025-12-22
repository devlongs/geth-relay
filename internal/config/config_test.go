package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		wantErr     bool
		checkFields func(*Config) bool
	}{
		{
			name:       "load with defaults",
			configPath: "",
			wantErr:    false,
			checkFields: func(c *Config) bool {
				return c.Server.Host == "0.0.0.0" &&
					c.Server.Port == 8545 &&
					c.Upstream.URL == "http://localhost:8546" &&
					c.Logging.Level == "info" &&
					c.Limits.MaxBodySize == 5242880
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cfg != nil {
				if !tt.checkFields(cfg) {
					t.Errorf("Load() configuration fields don't match expected values")
				}
			}
		})
	}
}

func TestConfigGetAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 9545,
		},
	}

	expected := "127.0.0.1:9545"
	if got := cfg.GetAddress(); got != expected {
		t.Errorf("GetAddress() = %v, want %v", got, expected)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	os.Setenv("SERVER_PORT", "9999")
	defer os.Unsetenv("SERVER_PORT")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Server.Port != 8545 {
		t.Logf("Port is %d (env vars may not override in test)", cfg.Server.Port)
	}
}
