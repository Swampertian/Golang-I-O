package logger

import (
	"log"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
)

var Log *slog.Logger

func Init() {
	if err := os.MkdirAll("/var/log/app", 0o755); err != nil {
		log.Fatalf("erro criando /var/log/app: %v", err)
	}

	f, err := os.OpenFile("/var/log/app/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalf("erro abrindo app.log: %v", err)
	}

	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	fileHandler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	multi := slogmulti.Fanout(stdoutHandler, fileHandler)

	Log = slog.New(multi)

	Log.Info("logger_initialized", "path", "/var/log/app/app.log")
}
