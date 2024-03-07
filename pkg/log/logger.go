package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Option represents an option for the logger.
type Option func(*Options)

// Options represents the options for the logger.
type Options struct {
	// CLI creates a logger that can be used in CLI applications.
	CLI bool
}

// WithCLI enables human-readable output.
func WithCLI() Option {
	return func(options *Options) {
		options.CLI = true
	}
}

// NewLogger creates a new logger with the default configuration.
func NewLogger(options ...Option) *zap.Logger {
	opts := &Options{}
	for _, option := range options {
		option(opts)
	}

	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if opts.CLI {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.DisableStacktrace = true
		config.DisableCaller = true
	}

	// Ensure that the logger is compatible with loki.
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.TimeKey = "timestamp"

	logger, err := config.Build()
	if err != nil {
		// Fall back to a simple JSON log line.
		fmt.Printf(`{"level":"error","timestamp":"%s","msg":"failed to create logger","error":"%s"}`, time.Now().Format(time.RFC3339), err)
		fmt.Println()
		os.Exit(1)
	}

	return logger
}

var (
	// singleton is the singleton logger instance.
	singleton *zap.Logger
)

// NewSingletonLogger creates a new logger with the
// specified options. All subsequent calls to this
// function will return the same logger instance.
func NewSingletonLogger(options ...Option) *zap.Logger {
	if singleton == nil {
		singleton = NewLogger(options...)
	}

	return singleton
}
