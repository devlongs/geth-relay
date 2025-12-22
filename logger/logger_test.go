package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		format    string
		wantErr   bool
		wantLevel zapcore.Level
	}{
		{
			name:      "info level json format",
			level:     "info",
			format:    "json",
			wantErr:   false,
			wantLevel: zapcore.InfoLevel,
		},
		{
			name:      "debug level console format",
			level:     "debug",
			format:    "console",
			wantErr:   false,
			wantLevel: zapcore.DebugLevel,
		},
		{
			name:      "warn level",
			level:     "warn",
			format:    "json",
			wantErr:   false,
			wantLevel: zapcore.WarnLevel,
		},
		{
			name:      "error level",
			level:     "error",
			format:    "json",
			wantErr:   false,
			wantLevel: zapcore.ErrorLevel,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			format:  "json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("New() returned nil logger")
			}
			if !tt.wantErr && logger != nil {
				logger.Sync()
			}
		})
	}
}
