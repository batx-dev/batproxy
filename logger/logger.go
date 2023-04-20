package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

type Options struct {
	// slog.Logger set for return a source parent logger
	*slog.Logger

	// slog.HandlerOptions set for slog options
	slog.HandlerOptions

	JsonHandler bool
}

func New(opts Options) *slog.Logger {
	if opts.Logger != nil {
		return opts.Logger
	}

	if &opts.HandlerOptions == nil {
		opts.HandlerOptions = slog.HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		}
	}

	if opts.JsonHandler {
		return slog.New(opts.NewJSONHandler(os.Stdout))
	}

	return slog.New(opts.NewTextHandler(os.Stdout))
}
