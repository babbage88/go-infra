package customlogger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"time"
)

type CustomLogger struct {
	LogFileName      string
	RotationInterval int16
	RotationTime     time.Time
	LoggingLevel     slog.Level
}

type CustomLoggerOption func(*CustomLogger)

// GetHostname retrieves and returns the hostname
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Info(fmt.Sprint(err))
		os.Exit(1)
	}

	return hostname
}

// GetBuildInfo retrieves and returns the build info
func GetBuildInfo() *debug.BuildInfo {
	buildInfo, _ := debug.ReadBuildInfo()
	return buildInfo
}

func NewCustomLogger(opts ...CustomLoggerOption) *CustomLogger {
	const (
		defaultLogFileName      = ""
		defaultRotationInterval = 24 * -1
		defaulLoggingLevel      = slog.LevelInfo
	)

	defaultRotationTime := time.Now().Local().Add(time.Hour * defaultRotationInterval)

	c := &CustomLogger{
		LogFileName:      defaultLogFileName,
		RotationInterval: defaultRotationInterval,
		RotationTime:     defaultRotationTime,
		LoggingLevel:     defaulLoggingLevel,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithLogFileName(logFileName string) CustomLoggerOption {
	return func(c *CustomLogger) {
		c.LogFileName = logFileName
	}
}

func WithRotationInterval(rotateint int16) CustomLoggerOption {
	return func(c *CustomLogger) {
		c.RotationInterval = rotateint
	}
}

func WithRotationTime(rotatetime time.Time) CustomLoggerOption {
	return func(c *CustomLogger) {
		c.RotationTime = rotatetime
	}
}

func WithLogLevel(loglev slog.Level) CustomLoggerOption {
	return func(c *CustomLogger) {
		c.LoggingLevel = loglev
	}
}

func HandleLogFile(config *CustomLogger) io.Writer {
	if config.LogFileName == "" {

		multi := io.MultiWriter(os.Stdout)
		return multi

	} else {
		file, err := os.OpenFile(config.LogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			if os.IsNotExist(err) {
				file, err = os.Create(config.LogFileName)
				if err != nil {
					fmt.Println("Error creating log file:", err)
					os.Exit(1)
				}
			} else {
				fmt.Println("Error opening log file:", err)
				os.Exit(1)
			}
		}

		multi := io.MultiWriter(os.Stdout, file)
		return multi
	}
}

// SetupLogger configures and returns a new logger based on the provided configuration
func SetupLogger(config *CustomLogger) *slog.Logger {

	multi := HandleLogFile(config)
	handler := slog.NewJSONHandler(multi, nil)

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(config.LoggingLevel)

	// Create a child logger with additional context
	curHost := GetHostname()
	buildInfo, _ := debug.ReadBuildInfo()

	child := logger.With(
		slog.String("Hostname", curHost),
		slog.Group("program_info",
			slog.Int("pid", os.Getpid()),
			slog.String("go_version", buildInfo.GoVersion),
		),
	)

	return child
}
