package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
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

	file, err := os.OpenFile("infra-api.log", os.O_RDWR|os.O_CREATE, 0644)
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

	file.Close()
}
