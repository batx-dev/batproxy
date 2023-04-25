package logger

import (
	"os"

	"github.com/batx-dev/batproxy"
	"golang.org/x/exp/slog"
)

type Logger struct {
	*slog.Logger
}

type Options struct {
	// slog.Logger set for return a source parent logger
	*slog.Logger

	// slog.HandlerOptions set for slog options
	slog.HandlerOptions

	// JsonHandler default false, means default log as `Text` format,
	// non-effective when *slog.Logger is not nil
	JsonHandler bool
}

func New(opts Options) *Logger {
	if opts.Logger != nil {
		return &Logger{opts.Logger}
	}

	if &opts.HandlerOptions == nil {
		opts.HandlerOptions = slog.HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		}
	}

	if opts.JsonHandler {
		return &Logger{slog.New(opts.NewJSONHandler(os.Stdout))}
	}

	return &Logger{slog.New(opts.NewTextHandler(os.Stdout))}
}

func logErr(logger *slog.Logger, msg string, err error) {
	code, message := batproxy.ErrorCode(err), batproxy.ErrorMessage(err)
	switch code {
	case "":
		logger.Info(msg)
	case batproxy.EINTERNAL:
		logger.Error(msg, "err", err)
	default:
		logger.Error(msg, "code", code, "err", message)
	}

}
