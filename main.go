package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"

	"go-infra/customlogger"
)

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Info(fmt.Sprint(err))
		os.Exit(1)
	}

	return hostname
}

func main() {
	customlogger.NewCustomLogger()

	file, err := os.OpenFile("infra-api.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if os.IsNotExist(err) {
		os.Create("infra-api.log")
	}
	multi := io.MultiWriter(os.Stdout, file)
	handler := slog.NewJSONHandler(multi, nil)

	logger := slog.New(handler)
	buildInfo, _ := debug.ReadBuildInfo()

	slog.SetDefault(logger)

	slog.SetLogLoggerLevel(slog.LevelInfo)

	servername := GetHostname()
	child := logger.With(
		slog.String("Hostname", servername),
		slog.Group("program_info",
			slog.Int("pid", os.Getpid()),
			slog.String("go_version", buildInfo.GoVersion),
		),
	)

	child.Info("Starting Infra tasks", slog.String("Action", "DNS Updated"))
	child.Info("test")
	logger.Info("Testin regular logger.")

	file.Close()
}
