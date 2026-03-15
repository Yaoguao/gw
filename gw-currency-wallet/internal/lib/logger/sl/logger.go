package sl

import (
	"gw-currency-wallet/internal/lib/logger/handlers/slogpretty"
	"io"
	"log/slog"
)

var logger *slog.Logger

func InitLogger(env string, w io.Writer) *slog.Logger {
	switch env {
	case "dev":
		logger = SetupPrettySlog(w, slog.LevelDebug)
	case "local":
		logger = SetupPrettySlog(w, slog.LevelDebug)
	default:
		logger = SetupPrettySlog(w, slog.LevelInfo)
	}

	return logger
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func SetupPrettySlog(w io.Writer, level slog.Level) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: level,
		},
	}
	handler := opts.NewPrettyHandler(w)

	return slog.New(handler)
}
